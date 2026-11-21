package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/middlewares"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestMiddlewares(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Middlewares Suite")
}

func histogramPoints(reader *sdkmetric.ManualReader) []metricdata.HistogramDataPoint[float64] {
	m, ok := findMetric(reader, "http.server.request.duration")
	if !ok {
		return nil
	}
	return m.Data.(metricdata.Histogram[float64]).DataPoints
}

func activeRequests(reader *sdkmetric.ManualReader) int64 {
	m, ok := findMetric(reader, "http.server.active_requests")
	Expect(ok).To(BeTrue(), "http.server.active_requests not recorded")
	var total int64
	for _, p := range m.Data.(metricdata.Sum[int64]).DataPoints {
		total += p.Value
	}
	return total
}

func findMetric(reader *sdkmetric.ManualReader, name string) (metricdata.Metrics, bool) {
	var rm metricdata.ResourceMetrics
	Expect(reader.Collect(context.Background(), &rm)).To(Succeed())
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == name {
				return m, true
			}
		}
	}
	return metricdata.Metrics{}, false
}

func attr(p metricdata.HistogramDataPoint[float64], key string) (attribute.Value, bool) {
	return p.Attributes.Value(attribute.Key(key))
}

var _ = Describe("Metrics", func() {
	var (
		reader *sdkmetric.ManualReader
		mux    *httpx.Mux
	)

	do := func(method, target string) int {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
		return rec.Code
	}

	BeforeEach(func() {
		reader = sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		DeferCleanup(provider.Shutdown, context.Background())

		metrics, err := middlewares.Metrics(provider)
		Expect(err).NotTo(HaveOccurred())

		mux = httpx.NewServeMux()
		mux.Use(metrics)
		mux.Use(middlewares.Recovery)
	})

	It("records duration with method, route and status", func() {
		mux.HandleFunc("GET /api/posts/{id}", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		Expect(do(http.MethodGet, "/api/posts/123")).To(Equal(http.StatusOK))

		points := histogramPoints(reader)
		Expect(points).To(HaveLen(1))
		p := points[0]
		Expect(p.Count).To(BeEquivalentTo(1))

		method, _ := attr(p, "http.request.method")
		Expect(method.AsString()).To(Equal("GET"))
		route, _ := attr(p, "http.route")
		Expect(route.AsString()).To(Equal("/api/posts/{id}"))
		status, _ := attr(p, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(200))
		scheme, _ := attr(p, "url.scheme")
		Expect(scheme.AsString()).To(Equal("http"))
		_, hasError := attr(p, "error.type")
		Expect(hasError).To(BeFalse())
	})

	It("groups different ids under one route", func() {
		mux.HandleFunc("GET /api/posts/{id}", func(http.ResponseWriter, *http.Request) {})

		for _, id := range []string{"1", "2", "3"} {
			do(http.MethodGet, "/api/posts/"+id)
		}

		points := histogramPoints(reader)
		Expect(points).To(HaveLen(1), "one time series, not one per id")
		Expect(points[0].Count).To(BeEquivalentTo(3))
	})

	It("uses second-based buckets", func() {
		mux.HandleFunc("GET /", func(http.ResponseWriter, *http.Request) {})
		do(http.MethodGet, "/")

		bounds := histogramPoints(reader)[0].Bounds
		Expect(bounds[0]).To(Equal(0.005))
		Expect(bounds[len(bounds)-1]).To(Equal(10.0))
	})

	It("records the status written by the handler", func() {
		mux.HandleFunc("POST /api/posts", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusTeapot) // superfluous, ignored by net/http
		})

		do(http.MethodPost, "/api/posts")

		status, _ := attr(histogramPoints(reader)[0], "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(201))
	})

	It("does not mark client errors as errors", func() {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})
		do(http.MethodGet, "/")

		_, hasError := attr(histogramPoints(reader)[0], "error.type")
		Expect(hasError).To(BeFalse())
	})

	It("marks server errors with error.type", func() {
		mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		})
		do(http.MethodGet, "/")

		errType, _ := attr(histogramPoints(reader)[0], "error.type")
		Expect(errType.AsString()).To(Equal("502"))
	})

	It("counts a recovered panic as 500", func() {
		mux.HandleFunc("GET /", func(http.ResponseWriter, *http.Request) { panic("boom") })

		Expect(do(http.MethodGet, "/")).To(Equal(http.StatusInternalServerError))

		status, _ := attr(histogramPoints(reader)[0], "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(500))
		Expect(activeRequests(reader)).To(BeZero())
	})

	It("counts an unrecovered panic as 500 and re-panics", func() {
		reader = sdkmetric.NewManualReader()
		provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		DeferCleanup(provider.Shutdown, context.Background())
		metrics, err := middlewares.Metrics(provider)
		Expect(err).NotTo(HaveOccurred())

		handler := metrics(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
		Expect(func() {
			handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
		}).To(PanicWith("boom"))

		status, _ := attr(histogramPoints(reader)[0], "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(500))
		Expect(activeRequests(reader)).To(BeZero())
	})

	It("maps non-standard methods to _OTHER", func() {
		mux.HandleFunc("/any", func(http.ResponseWriter, *http.Request) {})
		do("PURGE", "/any")

		method, _ := attr(histogramPoints(reader)[0], "http.request.method")
		Expect(method.AsString()).To(Equal("_OTHER"))
	})

	It("tracks active requests", func() {
		inside := make(chan struct{})
		release := make(chan struct{})
		mux.HandleFunc("GET /slow", func(http.ResponseWriter, *http.Request) {
			close(inside)
			<-release
		})

		done := make(chan struct{})
		go func() {
			defer close(done)
			do(http.MethodGet, "/slow")
		}()

		<-inside
		Expect(activeRequests(reader)).To(BeEquivalentTo(1))
		close(release)
		<-done
		Expect(activeRequests(reader)).To(BeZero())
	})

	It("keeps http.ResponseController working", func() {
		mux.HandleFunc("GET /stream", func(w http.ResponseWriter, _ *http.Request) {
			defer GinkgoRecover()
			Expect(http.NewResponseController(w).Flush()).To(Succeed())
		})

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/stream", nil))
		Expect(rec.Flushed).To(BeTrue())
	})
})
