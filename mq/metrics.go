package mq

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("mq")

type Metrics struct {
	TotalConsumedMessage metric.Int64Counter
	TotalSleep           metric.Int64Counter
	MsgError             metric.Int64Counter
}

func init() {
	core.Provides(newMetrics)
}

func newMetrics() *Metrics {

	otel.GetMeterProvider()
	metrics := &Metrics{}

	metrics.TotalConsumedMessage, _ = meter.Int64Counter(
		"mq.message",
		metric.WithUnit("{message}"),
		metric.WithDescription("Number of consumed message"),
	)

	metrics.TotalSleep, _ = meter.Int64Counter(
		"mq.consumer.sleep",
		metric.WithUnit("{sleep}"),
		metric.WithDescription("Number of sleep"),
	)

	metrics.MsgError, _ = meter.Int64Counter(
		"mq.message.error",
		metric.WithUnit("{message}"),
		metric.WithDescription("Number of message error"),
	)

	return metrics

}
