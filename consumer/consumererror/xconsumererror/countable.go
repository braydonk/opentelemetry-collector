// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package xconsumererror

type Countable interface {
	Count() int64
}

type countedError struct {
	inner error
	count int64
}

func (c *countedError) Error() string {
	if c.inner == nil {
		return "countable error"
	}
	return c.inner.Error()
}

func (c *countedError) Unwrap() error {
	return c.inner
}

func (c *countedError) Count() int64 {
	return c.count
}

// WrapCountableError wraps an error with a count. This is useful for errors that represent a specific number of items.
// For example, an error returned by a batch processing function might wrap the error with the number of items that failed to process.
func WrapCountableError(err error, count int64) error {
	return &countedError{
		inner: err,
		count: count,
	}
}

// ErrorAsCountable returns true if the error is countable, as well as a Countable interface.
func ErrorAsCountable(err error) (Countable, bool) {
	ce, ok := err.(Countable)
	return ce, ok
}
