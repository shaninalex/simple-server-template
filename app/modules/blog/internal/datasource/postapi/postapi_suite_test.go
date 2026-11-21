package postapi_test

import (
	"context"
	"testing"
	"time"
	"uuid"

	"simple-server-template/app/core/errs"
	"simple-server-template/app/modules/blog/internal/datasource/postapi"
	"simple-server-template/app/modules/blog/internal/domain"
	"simple-server-template/app/platform/httpclient"
	"simple-server-template/tests/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Post API Source Suite")
}

var _ = Describe("Source", func() {
	var source *postapi.Source

	BeforeEach(func() {
		remote := mocks.New()
		DeferCleanup(remote.Close)

		client, err := httpclient.New(httpclient.Config{BaseURL: remote.URL})
		Expect(err).NotTo(HaveOccurred())
		source = postapi.New(client)
	})

	Describe("List", func() {
		It("maps remote posts including category and author", func() {
			posts, err := source.List(context.Background(), 2, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(HaveLen(2))

			first := posts[0]
			Expect(first.Title).To(Equal("Getting Started with Go"))
			Expect(first.Category).To(Equal("it"))
			Expect(first.AuthorID).To(Equal(uuid.MustParse("550e8400-e29b-41d4-a716-446655440101")))
			Expect(first.CreatedAt.IsZero()).To(BeFalse())
		})

		It("applies the offset", func() {
			all, err := source.List(context.Background(), 2, 0)
			Expect(err).NotTo(HaveOccurred())

			shifted, err := source.List(context.Background(), 1, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(shifted).To(HaveLen(1))
			Expect(shifted[0].ID).To(Equal(all[1].ID))
		})
	})

	Describe("Create", func() {
		It("sends the post and returns the remote version", func() {
			post := domain.Post{
				AuthorID:  uuid.MustParse("550e8400-e29b-41d4-a716-446655440101"),
				Title:     "New post",
				Category:  "it",
				Content:   "Body",
				Published: time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC),
			}

			created, err := source.Create(context.Background(), post)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.ID).NotTo(Equal(uuid.UUID{}))
			Expect(created.Title).To(Equal(post.Title))
			Expect(created.Category).To(Equal(post.Category))
			Expect(created.AuthorID).To(Equal(post.AuthorID))
			Expect(created.Published.Equal(post.Published)).To(BeTrue())
		})

		It("makes the created post visible in the list", func() {
			created, err := source.Create(context.Background(), domain.Post{Title: "Listed"})
			Expect(err).NotTo(HaveOccurred())

			posts, err := source.List(context.Background(), 1000, 0)
			Expect(err).NotTo(HaveOccurred())
			Expect(posts).To(ContainElement(HaveField("ID", created.ID)))
		})
	})

	Describe("Get", func() {
		It("returns the post by id", func() {
			post, err := source.Get(context.Background(), uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"))
			Expect(err).NotTo(HaveOccurred())
			Expect(post.ID).To(Equal(uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")))
			Expect(post.Title).To(Equal("Getting Started with Go"))
			Expect(post.Category).To(Equal("it"))
		})

		It("returns a created post", func() {
			created, err := source.Create(context.Background(), domain.Post{Title: "Fresh"})
			Expect(err).NotTo(HaveOccurred())

			post, err := source.Get(context.Background(), created.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(post.Title).To(Equal("Fresh"))
		})

		It("reports a missing post as not found", func() {
			_, err := source.Get(context.Background(), uuid.New())
			Expect(errs.KindOf(err)).To(Equal(errs.KindNotFound))
		})
	})
})
