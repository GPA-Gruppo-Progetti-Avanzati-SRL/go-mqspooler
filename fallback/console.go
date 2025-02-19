package fallback

import (
	"context"
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
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

func (c *ConsoleHandler) Handle(ctx context.Context, payload string, err *core.ApplicationError) *core.ApplicationError {

	_, span := tracerConsole.Start(ctx, payload, trace.WithAttributes(
		attribute.String("fallback.type", ConsoleType),
		attribute.String("fallback.message.code", err.Code),
		attribute.String("fallback.message.error", err.Message),
		attribute.String("fallback.message.payload", payload),
	))

	defer span.End()
	log.Warn().Str("payload", payload).Err(err).Msgf("Fallback handle")
	if c.exitOnFailure {
		return core.BusinessErrorWithCodeAndMessage("Fallback", "Force exiting")
	}
	return nil

}

func NewConsoleHandler(cfg *Config) *ConsoleHandler {
	return &ConsoleHandler{
		exitOnFailure: cfg.ExitOnFailure,
	}

}
