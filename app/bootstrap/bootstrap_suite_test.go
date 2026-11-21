package bootstrap_test

import (
	"simple-server-template/app/bootstrap"
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
		It("creates an instance without errors", func() {
			remote := mocks.New()
			DeferCleanup(remote.Close)

			app, err := bootstrap.New(bootstrap.Config{
				PostsAPI: bootstrap.RemoteAPIConfig{BaseURL: remote.URL},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
			Expect(app.Core).NotTo(BeNil())
		})

		It("fails without posts API URL", func() {
			_, err := bootstrap.New(bootstrap.Config{})
			Expect(err).To(HaveOccurred())
		})
	})
})
