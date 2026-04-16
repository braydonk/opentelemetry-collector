// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package request // import "go.opentelemetry.io/collector/exporter/exporterhelper/internal/request"

import (
	"encoding"
	"fmt"
)

// TODO: Move this back to queuebatch when remove the circular dependency.

var (
	_ encoding.TextMarshaler   = (*SizerType)(nil)
	_ encoding.TextUnmarshaler = (*SizerType)(nil)
)

type SizerType struct {
	val string
}

const (
	sizerTypeBytes    = "bytes"
	sizerTypeItems    = "items"
	sizerTypeRequests = "requests"
	sizerTypeCompound = "compound"
)

var (
	SizerTypeBytes    = SizerType{val: sizerTypeBytes}
	SizerTypeItems    = SizerType{val: sizerTypeItems}
	SizerTypeRequests = SizerType{val: sizerTypeRequests}
	SizerTypeCompound = SizerType{val: sizerTypeCompound}
)

// UnmarshalText implements TextUnmarshaler interface.
func (s *SizerType) UnmarshalText(text []byte) error {
	switch str := string(text); str {
	case sizerTypeItems:
		*s = SizerTypeItems
	case sizerTypeBytes:
		*s = SizerTypeBytes
	case sizerTypeRequests:
		*s = SizerTypeRequests
	case sizerTypeCompound:
		*s = SizerTypeCompound
	default:
		return fmt.Errorf("invalid sizer: %q", str)
	}
	return nil
}

func (s *SizerType) MarshalText() ([]byte, error) {
	return []byte(s.val), nil
}

func (s *SizerType) String() string {
	return s.val
}

// Sizer is an interface that returns the size of the given element.
type Sizer interface {
	Sizeof(Request) int64
}

func NewSizer(sizerType SizerType) Sizer {
	switch sizerType {
	case SizerTypeBytes:
		return NewBytesSizer()
	case SizerTypeItems:
		return NewItemsSizer()
	default:
		return RequestsSizer{}
	}
}

// RequestsSizer is a Sizer implementation that returns the size of a queue element as one request.
type RequestsSizer struct{}

func (rs RequestsSizer) Sizeof(Request) int64 {
	return 1
}

type itemsSizer struct{}

func (itemsSizer) Sizeof(req Request) int64 {
	return int64(req.ItemsCount())
}

type bytesSizer struct{}

func (bytesSizer) Sizeof(req Request) int64 {
	return int64(req.BytesSize())
}

func NewItemsSizer() Sizer {
	return itemsSizer{}
}

func NewBytesSizer() Sizer {
	return bytesSizer{}
}

// BatchLimits defines limits for a batch in terms of requests, items, and bytes.
type BatchLimits struct {
	NumRequests int64 `mapstructure:"num_requests"`
	NumItems    int64 `mapstructure:"num_items"`
	NumBytes    int64 `mapstructure:"num_bytes"`
}
