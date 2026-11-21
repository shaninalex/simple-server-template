package telemetry_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"simple-server-template/app/platform/telemetry"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTelemetry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Telemetry Suite")
}

func newTelemetry(cfg telemetry.Config) *telemetry.Telemetry {
	t, err := telemetry.New(context.Background(), cfg)
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(t.Shutdown, context.Background())
	return t
}

func get(h http.Handler, path string) (int, string) {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	body, _ := io.ReadAll(rec.Body)
	return rec.Code, string(body)
}

var _ = Describe("Telemetry", func() {
	Describe("prometheus exporter", func() {
		var t *telemetry.Telemetry

		BeforeEach(func() {
			t = newTelemetry(telemetry.Config{
				ServiceName:    "test-service",
				ServiceVersion: "1.2.3",
				Exporter:       telemetry.ExporterPrometheus,
			})
		})

		It("exposes recorded metrics", func() {
			counter, err := t.MeterProvider().Meter("test").Int64Counter("test.requests")
			Expect(err).NotTo(HaveOccurred())
			counter.Add(context.Background(), 3)

			code, body := get(t.MetricsHandler(), "/metrics")
			Expect(code).To(Equal(http.StatusOK))
			Expect(body).To(MatchRegexp(`(?m)^test_requests_total\{.*\} 3$`))
		})

		It("exposes Go runtime metrics", func() {
			_, body := get(t.MetricsHandler(), "/metrics")
			Expect(body).To(ContainSubstring("go_goroutine_count"))
		})

		It("describes the service in target_info", func() {
			_, body := get(t.MetricsHandler(), "/metrics")
			Expect(body).To(MatchRegexp(`target_info\{.*service_name="test-service".*\}`))
			Expect(body).To(MatchRegexp(`target_info\{.*service_version="1.2.3".*\}`))
		})

		It("lets the environment override the service name", func() {
			GinkgoT().Setenv("OTEL_SERVICE_NAME", "from-env")
			t := newTelemetry(telemetry.Config{
				ServiceName: "from-config",
				Exporter:    telemetry.ExporterPrometheus,
			})

			_, body := get(t.MetricsHandler(), "/metrics")
			Expect(body).To(ContainSubstring(`service_name="from-env"`))
		})

		It("keeps instances independent", func() {
			other := newTelemetry(telemetry.Config{Exporter: telemetry.ExporterPrometheus})
			counter, err := other.MeterProvider().Meter("test").Int64Counter("only.other")
			Expect(err).NotTo(HaveOccurred())
			counter.Add(context.Background(), 1)

			_, body := get(t.MetricsHandler(), "/metrics")
			Expect(body).NotTo(ContainSubstring("only_other"))
		})
	})

	Describe("none exporter", func() {
		It("is the default and serves no metrics", func() {
			t := newTelemetry(telemetry.Config{})

			counter, err := t.MeterProvider().Meter("test").Int64Counter("test.requests")
			Expect(err).NotTo(HaveOccurred())
			counter.Add(context.Background(), 1)

			code, _ := get(t.MetricsHandler(), "/metrics")
			Expect(code).To(Equal(http.StatusNotFound))
		})
	})

	Describe("otlp exporter", func() {
		It("pushes metrics on shutdown", func() {
			received := make(chan string, 10)
			collector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				received <- r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			DeferCleanup(collector.Close)
			GinkgoT().Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", collector.URL)

			t, err := telemetry.New(context.Background(), telemetry.Config{Exporter: telemetry.ExporterOTLP})
			Expect(err).NotTo(HaveOccurred())

			counter, err := t.MeterProvider().Meter("test").Int64Counter("test.requests")
			Expect(err).NotTo(HaveOccurred())
			counter.Add(context.Background(), 1)

			Expect(t.Shutdown(context.Background())).To(Succeed())
			Eventually(received).Should(Receive(Equal("/v1/metrics")))
		})
	})

	It("rejects an unknown exporter", func() {
		_, err := telemetry.New(context.Background(), telemetry.Config{Exporter: "statsd"})
		Expect(err).To(MatchError(ContainSubstring(`unknown exporter "statsd"`)))
	})

	Describe("AdminHandler", func() {
		var admin http.Handler

		BeforeEach(func() {
			admin = newTelemetry(telemetry.Config{Exporter: telemetry.ExporterPrometheus}).AdminHandler()
		})

		It("serves /healthz", func() {
			code, _ := get(admin, "/healthz")
			Expect(code).To(Equal(http.StatusOK))
		})

		It("serves /metrics", func() {
			code, body := get(admin, "/metrics")
			Expect(code).To(Equal(http.StatusOK))
			Expect(body).To(ContainSubstring("target_info"))
		})
	})
})
