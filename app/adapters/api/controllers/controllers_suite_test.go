package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"simple-server-template/app/adapters/api"
	"simple-server-template/app/adapters/api/controllers"
	"simple-server-template/app/adapters/httpx"
	"simple-server-template/app/adapters/httpx/dto"
	"simple-server-template/app/bootstrap"
	"simple-server-template/app/core"
	"simple-server-template/tests/mocks"
	"strings"
	"testing"
	"uuid"

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

	Describe("POST /api/posts", func() {
		post := func(body string) *http.Response {
			ts := httptest.NewServer(controllers.PostCreate(resolve))
			DeferCleanup(ts.Close)

			res, err := http.Post(ts.URL, "application/json", strings.NewReader(body))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(res.Body.Close)
			return res
		}

		It("creates a post and returns it as DTO", func() {
			res := post(`{
				"author_id": "550e8400-e29b-41d4-a716-446655440101",
				"title": "New post",
				"category": "it",
				"content": "Body",
				"published": "2026-11-21T10:00:00Z"
			}`)
			Expect(res).To(HaveHTTPStatus(http.StatusCreated))

			var created dto.PostDTO
			Expect(json.NewDecoder(res.Body).Decode(&created)).To(Succeed())
			Expect(created.ID).NotTo(Equal(uuid.UUID{}))
			Expect(created.Title).To(Equal("New post"))
			Expect(created.Category).To(Equal("it"))
			Expect(created.AuthorID).To(Equal(uuid.MustParse("550e8400-e29b-41d4-a716-446655440101")))
			Expect(created.CreatedAt.IsZero()).To(BeFalse())
		})

		It("rejects malformed JSON", func() {
			Expect(post(`{"title":`)).To(HaveHTTPStatus(http.StatusBadRequest))
		})

		It("rejects unknown fields", func() {
			Expect(post(`{"title":"t","unknown":1}`)).To(HaveHTTPStatus(http.StatusBadRequest))
		})
	})

	Describe("GET /api/posts/{id}", func() {
		get := func(id string) *http.Response {
			router := httpx.NewServeMux()
			api.RegisterAPIControllers(router, resolve)
			ts := httptest.NewServer(router)
			DeferCleanup(ts.Close)

			res, err := http.Get(ts.URL + "/api/posts/" + id)
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(res.Body.Close)
			return res
		}

		It("returns the post as DTO", func() {
			res := get("550e8400-e29b-41d4-a716-446655440001")
			Expect(res).To(HaveHTTPStatus(http.StatusOK))

			var post dto.PostDTO
			Expect(json.NewDecoder(res.Body).Decode(&post)).To(Succeed())
			Expect(post.ID).To(Equal(uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")))
			Expect(post.Title).To(Equal("Getting Started with Go"))
			Expect(post.Category).To(Equal("it"))
		})

		It("returns 404 for an unknown post", func() {
			Expect(get(uuid.New().String())).To(HaveHTTPStatus(http.StatusNotFound))
		})

		It("returns 400 for an invalid id", func() {
			Expect(get("not-a-uuid")).To(HaveHTTPStatus(http.StatusBadRequest))
		})
	})

	Describe("when the posts API is down", func() {
		It("returns 502", func() {
			remote := mocks.New()
			remote.Close()

			app, err := bootstrap.New(bootstrap.Config{
				PostsAPI: bootstrap.RemoteAPIConfig{BaseURL: remote.URL},
			})
			Expect(err).NotTo(HaveOccurred())

			ts := httptest.NewServer(controllers.PostCreate(func() *core.App { return app.Core }))
			DeferCleanup(ts.Close)

			res, err := http.Post(ts.URL, "application/json", strings.NewReader(`{"title":"t"}`))
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(res.Body.Close)
			Expect(res).To(HaveHTTPStatus(http.StatusBadGateway))
		})
	})
})
