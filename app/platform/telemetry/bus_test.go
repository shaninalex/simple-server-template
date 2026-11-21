package telemetry_test

import (
	"context"
	"errors"

	"simple-server-template/app/core/bus"
	"simple-server-template/app/core/errs"
	"simple-server-template/app/platform/telemetry"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type createThing struct{ Name string }
type getThing struct{ ID int }
type thingCreated struct{}

func (thingCreated) EventName() string { return "thing.created" }

type busPoint = metricdata.HistogramDataPoint[float64]

func busPoints(reader *sdkmetric.ManualReader) []busPoint {
	var rm metricdata.ResourceMetrics
	Expect(reader.Collect(context.Background(), &rm)).To(Succeed())
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "bus.handler.duration" {
				return m.Data.(metricdata.Histogram[float64]).DataPoints
			}
		}
	}
	return nil
}

func busAttr(p busPoint, key string) string {
	v, ok := p.Attributes.Value(attribute.Key(key))
	if !ok {
		return ""
	}
	return v.AsString()
}

var _ = Describe("BusMetrics", func() {
	var (
		reader   *sdkmetric.ManualReader
		provider *sdkmetric.MeterProvider
	)

	middleware := func(kind telemetry.BusKind) bus.Middleware {
		mw, err := telemetry.BusMetrics(provider, kind)
		Expect(err).NotTo(HaveOccurred())
		return mw
	}

	BeforeEach(func() {
		reader = sdkmetric.NewManualReader()
		provider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		DeferCleanup(provider.Shutdown, context.Background())
	})

	Describe("commands", func() {
		var commands *bus.CommandBus

		register := func(h func(context.Context, createThing) (string, error)) {
			commands = bus.NewCommandBus(middleware(telemetry.BusCommand))
			Expect(bus.RegisterCommand(commands, h)).To(Succeed())
		}

		It("records a successful command", func() {
			register(func(context.Context, createThing) (string, error) { return "ok", nil })

			res, err := bus.Execute[createThing, string](context.Background(), commands, createThing{})
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(Equal("ok"))

			points := busPoints(reader)
			Expect(points).To(HaveLen(1))
			Expect(points[0].Count).To(BeEquivalentTo(1))
			Expect(busAttr(points[0], "bus.kind")).To(Equal("command"))
			Expect(busAttr(points[0], "bus.message")).To(Equal("telemetry_test.createThing"))
			Expect(busAttr(points[0], "error.type")).To(BeEmpty())
		})

		It("labels classified errors with their kind", func() {
			register(func(context.Context, createThing) (string, error) {
				return "", errs.NotFound("T1", "missing")
			})

			_, err := bus.Execute[createThing, string](context.Background(), commands, createThing{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindNotFound))

			Expect(busAttr(busPoints(reader)[0], "error.type")).To(Equal("not_found"))
		})

		It("labels plain errors as internal", func() {
			register(func(context.Context, createThing) (string, error) {
				return "", errors.New("boom")
			})

			_, err := bus.Execute[createThing, string](context.Background(), commands, createThing{})
			Expect(err).To(MatchError("boom"))

			Expect(busAttr(busPoints(reader)[0], "error.type")).To(Equal("internal"))
		})

		It("records a panic and lets it propagate", func() {
			register(func(context.Context, createThing) (string, error) { panic("boom") })

			Expect(func() {
				_, _ = bus.Execute[createThing, string](context.Background(), commands, createThing{})
			}).To(PanicWith("boom"))

			Expect(busAttr(busPoints(reader)[0], "error.type")).To(Equal("panic"))
		})

		It("keeps one series per message type", func() {
			register(func(context.Context, createThing) (string, error) { return "", nil })

			for _, name := range []string{"a", "b", "c"} {
				_, err := bus.Execute[createThing, string](context.Background(), commands, createThing{Name: name})
				Expect(err).NotTo(HaveOccurred())
			}

			points := busPoints(reader)
			Expect(points).To(HaveLen(1))
			Expect(points[0].Count).To(BeEquivalentTo(3))
		})
	})

	It("labels queries", func() {
		queries := bus.NewQueryBus(middleware(telemetry.BusQuery))
		Expect(bus.RegisterQuery(queries, func(context.Context, getThing) (int, error) { return 1, nil })).To(Succeed())

		_, err := bus.Ask[getThing, int](context.Background(), queries, getThing{})
		Expect(err).NotTo(HaveOccurred())

		p := busPoints(reader)[0]
		Expect(busAttr(p, "bus.kind")).To(Equal("query"))
		Expect(busAttr(p, "bus.message")).To(Equal("telemetry_test.getThing"))
	})

	It("records every event subscriber", func() {
		events := bus.NewEventBus(middleware(telemetry.BusEvent))
		Expect(bus.Subscribe(events, func(context.Context, thingCreated) error { return nil })).To(Succeed())
		Expect(bus.Subscribe(events, func(context.Context, thingCreated) error { panic("boom") })).To(Succeed())

		Expect(events.Publish(context.Background(), thingCreated{})).To(MatchError(ContainSubstring("panic")))

		var ok, panicked uint64
		for _, p := range busPoints(reader) {
			Expect(busAttr(p, "bus.kind")).To(Equal("event"))
			Expect(busAttr(p, "bus.message")).To(Equal("telemetry_test.thingCreated"))
			if busAttr(p, "error.type") == "panic" {
				panicked += p.Count
			} else {
				ok += p.Count
			}
		}
		Expect(ok).To(BeEquivalentTo(1))
		Expect(panicked).To(BeEquivalentTo(1))
	})

	It("names pointer messages by their element type", func() {
		events := bus.NewEventBus(middleware(telemetry.BusEvent))
		Expect(bus.Subscribe(events, func(context.Context, *thingCreated) error { return nil })).To(Succeed())

		Expect(events.Publish(context.Background(), &thingCreated{})).To(Succeed())

		Expect(busAttr(busPoints(reader)[0], "bus.message")).To(Equal("telemetry_test.thingCreated"))
	})
})
