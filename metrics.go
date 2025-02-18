package mqspooler

import (
	"github.com/GPA-Gruppo-Progetti-Avanzati-SRL/go-core-app"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var meter = otel.Meter("spooler")

type Metrics struct {
	MessageLatency metric.Int64Histogram
	//MessageInError metric.Int64Counter
}

func newMetrics() *Metrics {

	otel.GetMeterProvider()
	metrics := &Metrics{}

	metrics.MessageLatency, _ = meter.Int64Histogram(
		"spooler.processing.latency",
		metric.WithUnit("ms"),
		metric.WithDescription("Latency to process a single message. Only record that was able to process are recorder not one that failed"),
		// TODO bilancia il bucket
		//metric.WithExplicitBucketBoundaries(100, 2000),
	)
	//TODO handle inside fallback
	/*
		metrics.MessageInError, _ = meter.Int64Counter(
			"spooler.message.count.error",
			metric.WithUnit("{message}"),
			metric.WithDescription("Number message that go in error"),
			)*/

	return metrics

}

func init() {
	core.Provides(newMetrics)
}
