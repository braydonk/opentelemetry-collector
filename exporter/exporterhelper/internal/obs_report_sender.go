// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package internal // import "go.opentelemetry.io/collector/exporter/exporterhelper/internal"

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/consumer/consumererror/xconsumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper/internal/metadata"
	"go.opentelemetry.io/collector/exporter/exporterhelper/internal/queuebatch"
	"go.opentelemetry.io/collector/exporter/exporterhelper/internal/request"
	"go.opentelemetry.io/collector/exporter/exporterhelper/internal/sender"
	"go.opentelemetry.io/collector/pipeline"
)

const (
	// spanNameSep is duplicate between receiver and exporter.
	spanNameSep = "/"

	// ExporterKey used to identify exporters in metrics and traces.
	ExporterKey = "exporter"

	// DataTypeKey used to identify the data type in the queue size metric.
	DataTypeKey = "data_type"

	// ItemsSent used to track number of items sent by exporters.
	ItemsSent = "items.sent"
	// ItemsFailed used to track number of items that failed to be sent by exporters.
	ItemsFailed = "items.failed"
)

type obsReportSender[K request.Request] struct {
	component.StartFunc
	component.ShutdownFunc

	spanName        string
	tracer          trace.Tracer
	spanAttrs       trace.SpanStartEventOption
	metricAttr      metric.MeasurementOption
	itemsSentInst   metric.Int64Counter
	itemsFailedInst metric.Int64Counter
	next            sender.Sender[K]
}

func newObsReportSender[K request.Request](set exporter.Settings, signal pipeline.Signal, next sender.Sender[K]) (sender.Sender[K], error) {
	telemetryBuilder, err := metadata.NewTelemetryBuilder(set.TelemetrySettings)
	if err != nil {
		return nil, err
	}

	idStr := set.ID.String()
	expAttr := attribute.String(ExporterKey, idStr)

	or := &obsReportSender[K]{
		spanName:   ExporterKey + spanNameSep + idStr + spanNameSep + signal.String(),
		tracer:     metadata.Tracer(set.TelemetrySettings),
		spanAttrs:  trace.WithAttributes(expAttr, attribute.String(DataTypeKey, signal.String())),
		metricAttr: metric.WithAttributeSet(attribute.NewSet(expAttr)),
		next:       next,
	}

	switch signal {
	case pipeline.SignalTraces:
		or.itemsSentInst = telemetryBuilder.ExporterSentSpans
		or.itemsFailedInst = telemetryBuilder.ExporterSendFailedSpans

	case pipeline.SignalMetrics:
		or.itemsSentInst = telemetryBuilder.ExporterSentMetricPoints
		or.itemsFailedInst = telemetryBuilder.ExporterSendFailedMetricPoints

	case pipeline.SignalLogs:
		or.itemsSentInst = telemetryBuilder.ExporterSentLogRecords
		or.itemsFailedInst = telemetryBuilder.ExporterSendFailedLogRecords
	}

	return or, nil
}

func (ors *obsReportSender[K]) Send(ctx context.Context, req K) error {
	// Have to read the number of items before sending the request since the request can
	// be modified by the downstream components like the batcher.
	c := ors.startOp(ctx)
	items := req.ItemsCount()
	// Forward the data to the next consumer (this pusher is the next).
	err := ors.next.Send(c, req)
	ors.endOp(c, items, err)
	return err
}

// StartOp creates the span used to trace the operation. Returning
// the updated context and the created span.
func (ors *obsReportSender[K]) startOp(ctx context.Context) context.Context {
	ctx, _ = ors.tracer.Start(ctx,
		ors.spanName,
		ors.spanAttrs,
		trace.WithLinks(queuebatch.LinksFromContext(ctx)...))
	return ctx
}

// EndOp completes the export operation that was started with StartOp.
func (ors *obsReportSender[K]) endOp(ctx context.Context, numItems int, err error) {
	numSent, numFailedToSend := toNumItems(int64(numItems), err)

	// No metrics recorded for profiles.
	if ors.itemsSentInst != nil {
		ors.itemsSentInst.Add(ctx, numSent, ors.metricAttr)
	}
	// No metrics recorded for profiles.
	if ors.itemsFailedInst != nil {
		ors.itemsFailedInst.Add(ctx, numFailedToSend, ors.metricAttr)
	}

	span := trace.SpanFromContext(ctx)
	defer span.End()
	// End the span according to errors.
	if span.IsRecording() {
		span.SetAttributes(
			attribute.Int64(ItemsSent, numSent),
			attribute.Int64(ItemsFailed, numFailedToSend),
		)
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		}
	}
}

func toNumItems(numExportedItems int64, err error) (int64, int64) {
	if err == nil {
		return numExportedItems, 0
	}
	if partialErr, ok := xconsumererror.AsPartial(err); ok {
		numFailedItems := int64(partialErr.Failed())
		return numExportedItems - numFailedItems, numFailedItems
	}
	if numSignalFailed, ok := countSignalError(err); ok {
		return numExportedItems - numSignalFailed, numSignalFailed
	}
	return 0, numExportedItems
}

func countSignalError(err error) (int64, bool) {
	var logsErr consumererror.Logs
	if errors.As(err, &logsErr) {
		return int64(logsErr.Data().LogRecordCount()), true
	}
	var metricsErr consumererror.Metrics
	if errors.As(err, &metricsErr) {
		return int64(metricsErr.Data().DataPointCount()), true
	}
	var tracesErr consumererror.Traces
	if errors.As(err, &tracesErr) {
		return int64(tracesErr.Data().SpanCount()), true
	}
	return 0, false
}
