// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/log"

	"github.com/stretchr/testify/assert"
)

func BenchmarkProcessor(b *testing.B) {
	for _, tc := range []struct {
		name string
		f    func() Processor
	}{
		{
			name: "Simple",
			f: func() Processor {
				return NewSimpleProcessor(noopExporter{})
			},
		},
		{
			name: "Batch",
			f: func() Processor {
				return NewBatchProcessor(noopExporter{})
			},
		},
		{
			name: "SetTimestampSimple",
			f: func() Processor {
				return timestampDecorator{NewSimpleProcessor(noopExporter{})}
			},
		},
		{
			name: "SetTimestampBatch",
			f: func() Processor {
				return timestampDecorator{NewBatchProcessor(noopExporter{})}
			},
		},
		{
			name: "AddAttributesSimple",
			f: func() Processor {
				return attrAddDecorator{NewSimpleProcessor(noopExporter{})}
			},
		},
		{
			name: "AddAttributesBatch",
			f: func() Processor {
				return attrAddDecorator{NewBatchProcessor(noopExporter{})}
			},
		},
		{
			name: "SetAttributesSimple",
			f: func() Processor {
				return attrSetDecorator{NewSimpleProcessor(noopExporter{})}
			},
		},
		{
			name: "SetAttributesBatch",
			f: func() Processor {
				return attrSetDecorator{NewBatchProcessor(noopExporter{})}
			},
		},
	} {
		b.Run(tc.name, func(b *testing.B) {
			provider := NewLoggerProvider(WithProcessor(tc.f()))
			b.Cleanup(func() { assert.NoError(b, provider.Shutdown(context.Background())) })
			logger := provider.Logger(b.Name())

			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					r := log.Record{}
					r.SetBody(log.StringValue("message"))
					r.SetSeverity(log.SeverityInfo)
					r.AddAttributes(
						log.String("foo", "bar"),
						log.Float64("float", 3.14),
						log.Int("int", 123),
						log.Bool("bool", true),
					)
					logger.Emit(context.Background(), r)
				}
			})
		})
	}
}

type timestampDecorator struct {
	Processor
}

func (e timestampDecorator) OnEmit(ctx context.Context, r Record) error {
	r = r.Clone()
	r.SetObservedTimestamp(time.Date(1988, time.November, 17, 0, 0, 0, 0, time.UTC))
	return e.Processor.OnEmit(ctx, r)
}

type attrAddDecorator struct {
	Processor
}

func (e attrAddDecorator) OnEmit(ctx context.Context, r Record) error {
	r = r.Clone()
	r.AddAttributes(log.String("add", "me"))
	return e.Processor.OnEmit(ctx, r)
}

type attrSetDecorator struct {
	Processor
}

func (e attrSetDecorator) OnEmit(ctx context.Context, r Record) error {
	r = r.Clone()
	r.SetAttributes(log.String("replace", "me"))
	return e.Processor.OnEmit(ctx, r)
}
