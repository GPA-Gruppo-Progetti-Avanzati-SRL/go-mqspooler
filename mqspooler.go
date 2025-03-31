package mqspooler

import (
	"context"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-mqspooler/mq"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
	"time"
)

type Spooler struct {
	metrics         Metrics
	mq              *mq.MQConsumer
	closer          fx.Shutdowner
	shutdownChannel chan bool
	business        IBusiness
	died            bool
	currentSpan     trace.Span
	currentTime     time.Time
	fallback        IFallbackHandler
}
type IBusiness interface {
	ProcessMessage(ctx context.Context, message string) *core.ApplicationError
}
type IFallbackHandler interface {
	Handle(ctx context.Context, payload string, err *core.ApplicationError) *core.ApplicationError
}

var tracer = otel.Tracer("Spooler")

func NewSpooler(mq *mq.MQConsumer, metrics *Metrics,
	lc fx.Lifecycle, business IBusiness, closer fx.Shutdowner, fallback IFallbackHandler) *Spooler {

	spooler := &Spooler{
		mq:       mq,
		closer:   closer,
		metrics:  *metrics,
		business: business,
		fallback: fallback,
	}

	spooler.shutdownChannel = make(chan bool)
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go spooler.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info().Msg("Closing spooler")

			// Nel caso che abbia ricevuto una chiusura dal esterno per esempio un SIGTERM
			if !spooler.died {
				log.Info().Msg("Shutdown")
				spooler.shutdownChannel <- true
			}

			return nil
		},
	})

	return spooler
}

func (s *Spooler) Die(err *core.ApplicationError) {

	log.Error().Err(err).Msgf("Call Die %s - %s", err.Code, err.Message)
	if s.currentSpan != nil {
		s.currentSpan.SetStatus(codes.Error, err.Error())
		s.currentSpan.End()
	}
	s.died = true
	s.mq.Rollback()
	s.closer.Shutdown()

}

func (s *Spooler) Start() {

	defer func() {
		log.Trace().Msg("Shutting down")
		s.died = true
		s.closer.Shutdown()
	}()

	for {

		select {
		case _ = <-s.shutdownChannel:
			log.Info().Msg("received close from channel")
			return

		default:
			message := s.mq.ReadMessage()
			if message == nil {
				s.mq.Commit(true)
				continue
			}

			ctx, span := tracer.Start(context.Background(), "PROCESSING",
				trace.WithSpanKind(trace.SpanKindConsumer))
			s.currentSpan = span
			s.currentTime = time.Now()
			if errProcessing := s.business.ProcessMessage(ctx, *message); errProcessing != nil {
				if errProcessing.IsTechnicalError() {
					s.Die(errProcessing)
					return
				}
				s.metrics.MessageInFallback.Add(ctx, 1)
				if errFallback := s.fallback.Handle(ctx, *message, errProcessing); errFallback != nil {
					s.Die(errFallback)
					return
				}
			}
			s.mq.Commit(false)
			s.currentSpan.End()
			elapsed := time.Now().Sub(s.currentTime)
			s.metrics.MessageLatency.Record(context.Background(), elapsed.Milliseconds())
		}

	}

}
