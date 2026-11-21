package bootstrap_test

import (
	"context"
	"simple-server-template/app/bootstrap"
	"simple-server-template/app/platform/telemetry"
	"simple-server-template/tests/mocks"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBootstrap(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bootstrap Suite")
}

var _ = Describe("Bootstrap", func() {
	Describe("New", func() {
		var remoteURL string

		BeforeEach(func() {
			remote := mocks.New()
			DeferCleanup(remote.Close)
			remoteURL = remote.URL
		})

		It("creates an instance without errors", func() {
			app, err := bootstrap.New(bootstrap.Config{
				PostsAPI: bootstrap.RemoteAPIConfig{BaseURL: remoteURL},
			})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(app.Shutdown, context.Background())

			Expect(app).NotTo(BeNil())
			Expect(app.Core).NotTo(BeNil())
			Expect(app.Telemetry).NotTo(BeNil())
		})

		It("fails without posts API URL", func() {
			_, err := bootstrap.New(bootstrap.Config{})
			Expect(err).To(HaveOccurred())
		})

		It("fails with an unknown metrics exporter", func() {
			_, err := bootstrap.New(bootstrap.Config{
				PostsAPI:  bootstrap.RemoteAPIConfig{BaseURL: remoteURL},
				Telemetry: telemetry.Config{Exporter: "statsd"},
			})
			Expect(err).To(MatchError(ContainSubstring("telemetry")))
		})

		It("shuts down cleanly", func() {
			app, err := bootstrap.New(bootstrap.Config{
				PostsAPI:  bootstrap.RemoteAPIConfig{BaseURL: remoteURL},
				Telemetry: telemetry.Config{Exporter: telemetry.ExporterPrometheus},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(app.Shutdown(context.Background())).To(Succeed())
		})
	})

	Describe("LoadConfig", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("POSTS_API_URL", "http://posts.local")
		})

		It("requires POSTS_API_URL", func() {
			GinkgoT().Setenv("POSTS_API_URL", "")
			_, err := bootstrap.LoadConfig()
			Expect(err).To(MatchError(ContainSubstring("POSTS_API_URL")))
		})

		It("exports metrics to prometheus by default", func() {
			GinkgoT().Setenv("OTEL_METRICS_EXPORTER", "")
			cfg, err := bootstrap.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Telemetry.Exporter).To(Equal(telemetry.ExporterPrometheus))
			Expect(cfg.Telemetry.ServiceName).NotTo(BeEmpty())
			Expect(cfg.Telemetry.ServiceVersion).To(Equal(bootstrap.Version))
		})

		It("listens on :8080 and :9090 by default", func() {
			GinkgoT().Setenv("HTTP_ADDR", "")
			GinkgoT().Setenv("ADMIN_ADDR", "")
			cfg, err := bootstrap.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.HTTP.Addr).To(Equal(":8080"))
			Expect(cfg.HTTP.AdminAddr).To(Equal(":9090"))
		})

		It("takes addresses from HTTP_ADDR and ADMIN_ADDR", func() {
			GinkgoT().Setenv("HTTP_ADDR", ":8000")
			GinkgoT().Setenv("ADMIN_ADDR", "127.0.0.1:9000")
			cfg, err := bootstrap.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.HTTP.Addr).To(Equal(":8000"))
			Expect(cfg.HTTP.AdminAddr).To(Equal("127.0.0.1:9000"))
		})

		It("takes the exporter from OTEL_METRICS_EXPORTER", func() {
			GinkgoT().Setenv("OTEL_METRICS_EXPORTER", "otlp")
			cfg, err := bootstrap.LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Telemetry.Exporter).To(Equal(telemetry.ExporterOTLP))
		})
	})
})
