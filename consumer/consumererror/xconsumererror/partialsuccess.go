// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package xconsumererror

import (
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/consumer/consumererror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type partialSuccessError struct {
	inner error
}

func (p *partialSuccessError) Error() string {
	if p.inner == nil {
		return "partial success"
	}
	return fmt.Sprintf("partial success: %s", p.inner.Error())
}

func (p *partialSuccessError) Unwrap() error {
	return p.inner
}

func (p *partialSuccessError) GRPCStatus() *status.Status {
	if p.inner == nil {
		return status.New(codes.OK, "partial success")
	}
	return status.New(codes.OK, p.inner.Error())
}

// NewPartialSuccessError is
func NewPartialSuccessError(err error, count int64) error {
	innerErr := err
	if innerErr == nil {
		innerErr = errors.New("partial success")
	}
	if !consumererror.IsPermanent(innerErr) {
		innerErr = consumererror.NewPermanent(innerErr)
	}
	return WrapCountableError(
		&partialSuccessError{
			inner: innerErr,
		},
		count,
	)
}
