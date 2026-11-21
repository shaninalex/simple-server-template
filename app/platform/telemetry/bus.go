package telemetry

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"simple-server-template/app/core/bus"
	"simple-server-template/app/core/errs"
)

type BusKind string

const (
	BusCommand BusKind = "command"
	BusQuery   BusKind = "query"
	BusEvent   BusKind = "event"
)

const busMeterName = "simple-server-template/app/core/bus"

var busDurationBuckets = []float64{
	0.001, 0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10,
}

func BusMetrics(mp metric.MeterProvider, kind BusKind) (bus.Middleware, error) {
	duration, err := mp.Meter(busMeterName).Float64Histogram(
		"bus.handler.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Duration of command, query and event handlers."),
		metric.WithExplicitBucketBoundaries(busDurationBuckets...),
	)
	if err != nil {
		return nil, fmt.Errorf("bus metrics: %w", err)
	}

	return func(next bus.Invoke) bus.Invoke {
		return func(ctx context.Context, msg any) (res any, err error) {
			start := time.Now()
			attrs := []attribute.KeyValue{
				attribute.String("bus.kind", string(kind)),
				attribute.String("bus.message", messageName(msg)),
			}

			defer func() {
				p := recover()
				switch {
				case p != nil:
					attrs = append(attrs, attribute.String("error.type", "panic"))
				case err != nil:
					attrs = append(attrs, attribute.String("error.type", errs.KindOf(err).String()))
				}
				duration.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(attrs...))
				if p != nil {
					panic(p)
				}
			}()

			return next(ctx, msg)
		}
	}, nil
}

func messageName(msg any) string {
	t := reflect.TypeOf(msg)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil {
		return "nil"
	}
	return t.String()
}
