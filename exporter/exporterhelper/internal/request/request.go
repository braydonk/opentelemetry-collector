// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package request // import "go.opentelemetry.io/collector/exporter/exporterhelper/internal/request"

import (
	"context"
)

// Request represents a single request that can be sent to an external endpoint.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type Request interface {
	// ItemsCount returns a number of basic items in the request where item is the smallest piece of data that can be
	// sent. For example, for OTLP exporter, this value represents the number of spans,
	// metric data points or log records.
	ItemsCount() int
	// MergeSplit is a function that merge and/or splits this request with another one into multiple requests based on the
	// configured limit provided in MaxSizeConfig.
	// MergeSplit does not split if all fields in MaxSizeConfig are not initialized (zero).
	// All the returned requests MUST have a number of items that does not exceed the maximum number of items.
	// Size of the last returned request MUST be less or equal than the size of any other returned request.
	// The original request MUST not be mutated if error is returned after mutation or if the exporter is
	// marked as not mutable. The length of the returned slice MUST not be 0.
	// Experimental: This API is at the early stage of development and may change without backward compatibility
	// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
	MergeSplit(context.Context, int, SizerType, Request) ([]Request, error)
}

// ErrorHandler is an optional interface that can be implemented by Request to provide a way handle partial
// temporary failures. For example, if some items failed to process and can be retried, this interface allows to
// return a new Request that contains the items left to be sent. Otherwise, the original Request should be returned.
// If not implemented, the original Request will be returned assuming the error is applied to the whole Request.
// Experimental: This API is at the early stage of development and may change without backward compatibility
// until https://github.com/open-telemetry/opentelemetry-collector/issues/8122 is resolved.
type ErrorHandler interface {
	Request
	// OnError returns a new Request may contain the items left to be sent if some items failed to process and can be retried.
	// Otherwise, it should return the original Request.
	OnError(error) Request
}
