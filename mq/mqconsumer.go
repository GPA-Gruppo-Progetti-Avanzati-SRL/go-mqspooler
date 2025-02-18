package mq

import (
	"context"
	"github.com/ibm-messaging/mq-golang-jms20/jms20subset"
	"github.com/ibm-messaging/mq-golang-jms20/mqjms"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"
	"golang.org/x/text/encoding/charmap"
	"time"
)

type MQConsumer struct {
	TransactedContext         jms20subset.JMSContext
	Consumer                  jms20subset.JMSConsumer
	Config                    *Config
	CurrentTransactionMessage int
	metrics                   *Metrics
}

func NewMQConsumer(cfg *Config, lc fx.Lifecycle, metrics *Metrics) *MQConsumer {
	jmsConsumer := new(MQConsumer)
	jmsConsumer.Config = cfg
	jmsConsumer.CurrentTransactionMessage = 0
	jmsConsumer.metrics = metrics

	cf := mqjms.ConnectionFactoryImpl{
		QMName:      cfg.QueueManager,
		Hostname:    cfg.Hostname,
		PortNumber:  cfg.Port,
		ChannelName: cfg.Channel,
		UserName:    cfg.UserName,
		Password:    cfg.Password,

		TLSCipherSpec:    cfg.TLSCipherSpec,
		TLSClientAuth:    cfg.TLSClientAuth,
		KeyRepository:    cfg.KeyRepository,
		CertificateLabel: cfg.CertificateLabel,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			jmscontext, errCtx := cf.CreateContext()

			if errCtx != nil {
				log.Err(errCtx.GetLinkedError()).Msg("Errore connessione MQ")
				return errCtx

			}
			defer jmscontext.Close()
			queue := jmscontext.CreateQueue(cfg.QueueName)
			transactedContext, errCtx := cf.CreateContextWithSessionMode(jms20subset.JMSContextSESSIONTRANSACTED)
			jmsConsumer.TransactedContext = transactedContext
			consumer, conErr := transactedContext.CreateConsumer(queue)
			if conErr != nil {
				return conErr
			}
			jmsConsumer.Consumer = consumer
			return nil

		},
		OnStop: func(ctx context.Context) error {

			log.Info().Msg("Closing mq consumer")
			jmsConsumer.Consumer.Close()
			jmsConsumer.TransactedContext.Close()
			return nil
		},
	})

	return jmsConsumer

}

func (mqconsumer *MQConsumer) ReadMessageString() *[]byte {
	recBodyS, errJMS := mqconsumer.Consumer.ReceiveStringBodyNoWait()
	if errJMS != nil {
		log.Fatal().Msg("Jms Error : " + errJMS.GetErrorCode() + " : " + errJMS.GetReason())
	}
	if recBodyS == nil {
		return nil
	}
	data := []byte(*recBodyS)
	return &data

}
func (mqconsumer *MQConsumer) ReadMessageBytes() *[]byte {
	recBodyB, errJMS := mqconsumer.Consumer.ReceiveBytesBodyNoWait()
	if errJMS != nil {
		log.Fatal().Msg("Jms Error : " + errJMS.GetErrorCode() + " : " + errJMS.GetReason())
	}

	return recBodyB
}
func (mqconsumer *MQConsumer) ReadMessage() *string {

	var rcvBody *[]byte
	switch mqconsumer.Config.Mode {
	case "BINARY":
		rcvBody = mqconsumer.ReadMessageBytes()
		break
	case "STRING":
		rcvBody = mqconsumer.ReadMessageString()
		break
	default:
		log.Fatal().Msgf("Mode %s not managed", mqconsumer.Config.Mode)
	}

	if rcvBody != nil {
		mqconsumer.CurrentTransactionMessage++
		mqconsumer.metrics.TotalConsumedMessage.Add(context.Background(), 1)
		if !mqconsumer.CheckLength(rcvBody) {
			return nil
		}
		var input string
		if mqconsumer.Config.DecodeBody {
			input = decodeBytes(*rcvBody)
		}
		log.Trace().Msg("|" + input + "|")
		return &input

	} else {

		log.Trace().Msg("Sleeping...")
		mqconsumer.metrics.TotalSleep.Add(context.Background(), 1)
		time.Sleep(mqconsumer.Config.SleepNoMessage)
	}

	return nil

}

func (mqconsumer *MQConsumer) Commit(flush bool) {
	if flush || mqconsumer.CurrentTransactionMessage > mqconsumer.Config.MaxTransactionMessage {
		mqconsumer.TransactedContext.Commit()
		mqconsumer.CurrentTransactionMessage = 0
	}
}

func (mqconsumer *MQConsumer) Rollback() {
	mqconsumer.TransactedContext.Rollback()
	mqconsumer.CurrentTransactionMessage = 0
}

func decodeBytes(data []byte) string {
	decoder := charmap.CodePage037.NewDecoder()
	output, err := decoder.Bytes(data)
	if err != nil {
		log.Error().Msgf("Error %v", err)
	}
	r := string(output[:])
	return r
}

func (mqconsumer *MQConsumer) CheckLength(rcvBody *[]byte) bool {
	if mqconsumer.Config.Length == 0 {
		return true
	}
	lenbody := len(*rcvBody)

	if lenbody != mqconsumer.Config.Length {
		log.Warn().Msgf("Lunghezza messaggio non corretta %d\n", lenbody)
		mqconsumer.TransactedContext.Commit()
		mqconsumer.metrics.MsgError.Add(context.Background(), 1)
		return false
	}
	return true
}
