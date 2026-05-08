// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package xconsumererror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewPartialSuccessError(t *testing.T) {
	err := errors.New("some error")
	var count int64 = 42

	pErr := NewPartialSuccessError(err, count)
	assert.Equal(t, "partial success: Permanent error: some error", pErr.Error())

	// Test Countable behaviour
	c, ok := ErrorAsCountable(pErr)
	require.True(t, ok)
	assert.Equal(t, count, c.Count())

	// Test Permanent behaviour
	require.True(t, consumererror.IsPermanent(pErr), "should be detectable as permanent error")

	// Test Unwrap behaviour
	partialErrUnwrapped := errors.Unwrap(pErr) // Unwrap from countableError
	require.NotNil(t, partialErrUnwrapped)
	innerErr := errors.Unwrap(partialErrUnwrapped) // Unwrap from partialSuccessError
	expectedInnerErr := consumererror.NewPermanent(err)
	assert.Equal(t, expectedInnerErr, innerErr)

	// Test gRPC Status from error
	st, ok := status.FromError(pErr)
	if assert.True(t, ok, "status.FromError should successfully extract status") {
		assert.Equal(t, codes.OK, st.Code())
		assert.Equal(t, "partial success: Permanent error: some error", st.Message())
	}
}

func TestNewPartialSuccessError_NilInner(t *testing.T) {
	pErr := NewPartialSuccessError(nil, 10)
	assert.Equal(t, "partial success: Permanent error: partial success", pErr.Error())

	c, ok := ErrorAsCountable(pErr)
	require.True(t, ok)
	assert.Equal(t, int64(10), c.Count())

	st, ok := status.FromError(pErr)
	if assert.True(t, ok, "status.FromError should successfully extract status") {
		assert.Equal(t, codes.OK, st.Code())
		assert.Equal(t, "partial success: Permanent error: partial success", st.Message())
	}
}
