package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"simple-server-template/app/adapters/api/controllers"
	"simple-server-template/app/adapters/httpx/dto"
	"simple-server-template/app/bootstrap"
	"simple-server-template/app/core"
	"simple-server-template/tests/mocks"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controllers Suite")
}

var _ = Describe("Posts API endpoint", func() {
	var resolve core.Resolve

	BeforeEach(func() {
		remote := mocks.New()
		DeferCleanup(remote.Close)

		app, err := bootstrap.New(bootstrap.Config{
			PostsAPI: bootstrap.RemoteAPIConfig{BaseURL: remote.URL},
		})
		Expect(err).NotTo(HaveOccurred())
		resolve = func() *core.App { return app.Core }
	})

	Describe("GET /api/posts", func() {
		It("returns 200", func() {
			ts := httptest.NewServer(controllers.PostList(resolve))
			DeferCleanup(ts.Close)

			res, err := http.Get(ts.URL)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(res.Body.Close)

			Expect(res).To(HaveHTTPStatus(http.StatusOK))

			var posts []dto.PostDTO
			Expect(json.NewDecoder(res.Body).Decode(&posts)).To(Succeed())
			Expect(posts).NotTo(BeEmpty())
			Expect(len(posts)).To(BeNumerically("<=", 10))
			Expect(posts[0].Title).To(Equal("Getting Started with Go"))
		})
	})
})
