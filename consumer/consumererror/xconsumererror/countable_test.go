// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package xconsumererror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapCountableError(t *testing.T) {
	errMsg := "test error"
	inner := errors.New(errMsg)
	var count int64 = 5

	err := WrapCountableError(inner, count)
	require.NotNil(t, err)

	assert.Equal(t, errMsg, err.Error())

	c, ok := ErrorAsCountable(err)
	require.True(t, ok)
	assert.Equal(t, count, c.Count())

	assert.Equal(t, inner, errors.Unwrap(err))
}

func TestWrapCountableError_NilInner(t *testing.T) {
	var count int64 = 10

	err := WrapCountableError(nil, count)
	require.NotNil(t, err)

	assert.Equal(t, "countable error", err.Error())

	c, ok := ErrorAsCountable(err)
	require.True(t, ok)
	assert.Equal(t, count, c.Count())

	assert.Nil(t, errors.Unwrap(err))
}

func TestErrorAsCountable_False(t *testing.T) {
	c, ok := ErrorAsCountable(errors.New("non-countable"))
	assert.False(t, ok)
	assert.Nil(t, c)

	c, ok = ErrorAsCountable(nil)
	assert.False(t, ok)
	assert.Nil(t, c)
}
