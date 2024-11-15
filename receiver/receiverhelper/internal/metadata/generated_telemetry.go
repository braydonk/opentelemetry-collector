// Code generated by mdatagen. DO NOT EDIT.

package metadata

import (
	"errors"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configtelemetry"
)

// Deprecated: [v0.108.0] use LeveledMeter instead.
func Meter(settings component.TelemetrySettings) metric.Meter {
	return settings.MeterProvider.Meter("go.opentelemetry.io/collector/receiver/receiverhelper")
}

func LeveledMeter(settings component.TelemetrySettings, level configtelemetry.Level) metric.Meter {
	return settings.LeveledMeterProvider(level).Meter("go.opentelemetry.io/collector/receiver/receiverhelper")
}

func Tracer(settings component.TelemetrySettings) trace.Tracer {
	return settings.TracerProvider.Tracer("go.opentelemetry.io/collector/receiver/receiverhelper")
}

// TelemetryBuilder provides an interface for components to report telemetry
// as defined in metadata and user config.
type TelemetryBuilder struct {
	meter                        metric.Meter
	ReceiverAcceptedLogRecords   metric.Int64Counter
	ReceiverAcceptedMetricPoints metric.Int64Counter
	ReceiverAcceptedSpans        metric.Int64Counter
	ReceiverRefusedLogRecords    metric.Int64Counter
	ReceiverRefusedMetricPoints  metric.Int64Counter
	ReceiverRefusedSpans         metric.Int64Counter
	meters                       map[configtelemetry.Level]metric.Meter
}

// TelemetryBuilderOption applies changes to default builder.
type TelemetryBuilderOption interface {
	apply(*TelemetryBuilder)
}

type telemetryBuilderOptionFunc func(mb *TelemetryBuilder)

func (tbof telemetryBuilderOptionFunc) apply(mb *TelemetryBuilder) {
	tbof(mb)
}

// NewTelemetryBuilder provides a struct with methods to update all internal telemetry
// for a component
func NewTelemetryBuilder(settings component.TelemetrySettings, options ...TelemetryBuilderOption) (*TelemetryBuilder, error) {
	builder := TelemetryBuilder{meters: map[configtelemetry.Level]metric.Meter{}}
	for _, op := range options {
		op.apply(&builder)
	}
	if settings.LeveledMeterProvider == nil {
		return nil, errors.New("TelemetrySettings must have a LeveledMeterProvider")
	}
	builder.meters[configtelemetry.LevelBasic] = LeveledMeter(settings, configtelemetry.LevelBasic)
	var err, errs error
	builder.ReceiverAcceptedLogRecords, err = builder.meters[configtelemetry.LevelBasic].Int64Counter(
		"otelcol_receiver_accepted_log_records",
		metric.WithDescription("Number of log records successfully pushed into the pipeline."),
		metric.WithUnit("{records}"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverAcceptedMetricPoints, err = builder.meters[configtelemetry.LevelBasic].Int64Counter(
		"otelcol_receiver_accepted_metric_points",
		metric.WithDescription("Number of metric points successfully pushed into the pipeline."),
		metric.WithUnit("{datapoints}"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverAcceptedSpans, err = builder.meters[configtelemetry.LevelBasic].Int64Counter(
		"otelcol_receiver_accepted_spans",
		metric.WithDescription("Number of spans successfully pushed into the pipeline."),
		metric.WithUnit("{spans}"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverRefusedLogRecords, err = builder.meters[configtelemetry.LevelBasic].Int64Counter(
		"otelcol_receiver_refused_log_records",
		metric.WithDescription("Number of log records that could not be pushed into the pipeline."),
		metric.WithUnit("{records}"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverRefusedMetricPoints, err = builder.meters[configtelemetry.LevelBasic].Int64Counter(
		"otelcol_receiver_refused_metric_points",
		metric.WithDescription("Number of metric points that could not be pushed into the pipeline."),
		metric.WithUnit("{datapoints}"),
	)
	errs = errors.Join(errs, err)
	builder.ReceiverRefusedSpans, err = builder.meters[configtelemetry.LevelBasic].Int64Counter(
		"otelcol_receiver_refused_spans",
		metric.WithDescription("Number of spans that could not be pushed into the pipeline."),
		metric.WithUnit("{spans}"),
	)
	errs = errors.Join(errs, err)
	return &builder, errs
}
