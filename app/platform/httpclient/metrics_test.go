package httpclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"simple-server-template/app/platform/httpclient"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type clientPoint = metricdata.HistogramDataPoint[float64]

func clientPoints(reader *sdkmetric.ManualReader) []clientPoint {
	var rm metricdata.ResourceMetrics
	Expect(reader.Collect(context.Background(), &rm)).To(Succeed())
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == "http.client.request.duration" {
				return m.Data.(metricdata.Histogram[float64]).DataPoints
			}
		}
	}
	return nil
}

func value(p clientPoint, key string) (attribute.Value, bool) {
	return p.Attributes.Value(attribute.Key(key))
}

func pointWith(points []clientPoint, key string, want int64) clientPoint {
	for _, p := range points {
		if v, ok := value(p, key); ok && v.AsInt64() == want {
			return p
		}
	}
	Fail("no data point with " + key + "=" + strconv.FormatInt(want, 10))
	return clientPoint{}
}

var _ = Describe("Client metrics", func() {
	var (
		reader   *sdkmetric.ManualReader
		provider *sdkmetric.MeterProvider
	)

	client := func(baseURL string, retries int, timeout time.Duration) *httpclient.Client {
		c, err := httpclient.New(httpclient.Config{
			BaseURL:       baseURL,
			Timeout:       timeout,
			Retries:       retries,
			RetryBackoff:  time.Millisecond,
			MeterProvider: provider,
		})
		Expect(err).NotTo(HaveOccurred())
		return c
	}

	BeforeEach(func() {
		reader = sdkmetric.NewManualReader()
		provider = sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
		DeferCleanup(provider.Shutdown, context.Background())
	})

	It("records a successful request", func() {
		srv := serve(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(payload{Name: "ok"})
		})
		u, _ := url.Parse(srv.URL)
		port, _ := strconv.Atoi(u.Port())

		Expect(client(srv.URL, 0, time.Second).GetJSON(context.Background(), "/posts/1", nil, &payload{})).To(Succeed())

		points := clientPoints(reader)
		Expect(points).To(HaveLen(1))
		p := points[0]
		Expect(p.Count).To(BeEquivalentTo(1))

		method, _ := value(p, "http.request.method")
		Expect(method.AsString()).To(Equal("GET"))
		host, _ := value(p, "server.address")
		Expect(host.AsString()).To(Equal(u.Hostname()))
		srvPort, _ := value(p, "server.port")
		Expect(srvPort.AsInt64()).To(BeEquivalentTo(port))
		status, _ := value(p, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(200))
		scheme, _ := value(p, "url.scheme")
		Expect(scheme.AsString()).To(Equal("http"))

		_, hasError := value(p, "error.type")
		Expect(hasError).To(BeFalse())
		_, hasResend := value(p, "http.request.resend_count")
		Expect(hasResend).To(BeFalse())
		_, hasPath := value(p, "url.full")
		Expect(hasPath).To(BeFalse())
	})

	It("records every retry attempt with its resend count", func() {
		var calls atomic.Int32
		srv := serve(func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_ = json.NewEncoder(w).Encode(payload{})
		})

		Expect(client(srv.URL, 2, time.Second).GetJSON(context.Background(), "/posts", nil, &payload{})).To(Succeed())

		points := clientPoints(reader)
		Expect(points).To(HaveLen(3))

		var first clientPoint
		for _, p := range points {
			if _, resent := value(p, "http.request.resend_count"); !resent {
				first = p
			}
		}
		status, _ := value(first, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(503))
		errType, _ := value(first, "error.type")
		Expect(errType.AsString()).To(Equal("503"))

		second := pointWith(points, "http.request.resend_count", 1)
		status, _ = value(second, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(503))

		third := pointWith(points, "http.request.resend_count", 2)
		status, _ = value(third, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(200))
	})

	It("marks client errors with error.type", func() {
		srv := serve(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		Expect(client(srv.URL, 0, time.Second).GetJSON(context.Background(), "/posts/1", nil, &payload{})).NotTo(Succeed())

		errType, _ := value(clientPoints(reader)[0], "error.type")
		Expect(errType.AsString()).To(Equal("404"))
	})

	It("marks an undecodable response", func() {
		srv := serve(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("not json"))
		})

		Expect(client(srv.URL, 0, time.Second).GetJSON(context.Background(), "/posts", nil, &payload{})).NotTo(Succeed())

		p := clientPoints(reader)[0]
		errType, _ := value(p, "error.type")
		Expect(errType.AsString()).To(Equal("decode"))
		status, _ := value(p, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(200))
	})

	It("marks a timeout", func() {
		srv := serve(func(w http.ResponseWriter, r *http.Request) {
			select {
			case <-r.Context().Done():
			case <-time.After(time.Second):
			}
		})

		Expect(client(srv.URL, 0, 20*time.Millisecond).GetJSON(context.Background(), "/posts", nil, &payload{})).NotTo(Succeed())

		p := clientPoints(reader)[0]
		errType, _ := value(p, "error.type")
		Expect(errType.AsString()).To(Equal("timeout"))
		_, hasStatus := value(p, "http.response.status_code")
		Expect(hasStatus).To(BeFalse())
	})

	It("marks an unreachable server", func() {
		srv := httptest.NewServer(http.NotFoundHandler())
		srv.Close()

		Expect(client(srv.URL, 0, time.Second).GetJSON(context.Background(), "/posts", nil, &payload{})).NotTo(Succeed())

		errType, _ := value(clientPoints(reader)[0], "error.type")
		Expect(errType.AsString()).To(Equal("_OTHER"))
	})

	It("marks a request cancelled by the caller", func() {
		started := make(chan struct{})
		srv := serve(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-r.Context().Done()
		})

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-started
			cancel()
		}()
		Expect(client(srv.URL, 2, time.Second).GetJSON(ctx, "/posts", nil, &payload{})).NotTo(Succeed())

		points := clientPoints(reader)
		Expect(points).To(HaveLen(1))
		errType, _ := value(points[0], "error.type")
		Expect(errType.AsString()).To(Equal("canceled"))
	})

	It("records POST requests", func() {
		srv := serve(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(payload{})
		})

		Expect(client(srv.URL, 0, time.Second).Post(context.Background(), "/posts", payload{}, &payload{})).To(Succeed())

		p := clientPoints(reader)[0]
		method, _ := value(p, "http.request.method")
		Expect(method.AsString()).To(Equal("POST"))
		status, _ := value(p, "http.response.status_code")
		Expect(status.AsInt64()).To(BeEquivalentTo(201))
	})

	It("works without a meter provider", func() {
		srv := serve(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(payload{})
		})
		c, err := httpclient.New(httpclient.Config{BaseURL: srv.URL})
		Expect(err).NotTo(HaveOccurred())
		Expect(c.GetJSON(context.Background(), "/posts", nil, &payload{})).To(Succeed())
	})
})
