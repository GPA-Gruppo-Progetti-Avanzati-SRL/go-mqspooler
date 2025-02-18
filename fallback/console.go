package fallback

import (
	"context"
	"errors"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const ConsoleType string = "CONSOLE"

var tracerConsole = otel.Tracer("console.fallback")

type ConsoleHandler struct {
	exitOnFailure bool
}

func (c *ConsoleHandler) Handle(ctx context.Context, step, id, payload string, err error) error {

	_, span := tracerConsole.Start(ctx, step, trace.WithAttributes(
		attribute.String("fallback.type", ConsoleType),
		attribute.String("fallback.message.error", err.Error()),
		attribute.String("fallback.message.payload", payload),
	))

	defer span.End()
	log.Warn().Str("id", id).Str("step", step).Str("payload", payload).Err(err).Msgf("Fallback handle")
	if c.exitOnFailure {
		return errors.New("Force exiting")
	}
	return nil

}

func NewConsoleHandler(cfg *Config) *ConsoleHandler {
	return &ConsoleHandler{
		exitOnFailure: cfg.ExitOnFailure,
	}

}
