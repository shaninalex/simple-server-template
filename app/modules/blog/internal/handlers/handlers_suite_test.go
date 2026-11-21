package handlers_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"uuid"

	"simple-server-template/app/modules/blog/contract"
	"simple-server-template/app/modules/blog/internal/domain"
	"simple-server-template/app/modules/blog/internal/handlers"
	"simple-server-template/app/platform/clock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blog Handlers Suite")
}

// fakeSource records what the handler sends and returns a canned result.
type fakeSource struct {
	created domain.Post
	result  domain.Post
	err     error
}

func (f *fakeSource) List(context.Context, int, int) ([]domain.Post, error) {
	return nil, nil
}

func (f *fakeSource) Create(_ context.Context, p domain.Post) (domain.Post, error) {
	f.created = p
	return f.result, f.err
}

var _ = Describe("CreatePost", func() {
	var (
		source  *fakeSource
		clk     *clock.Mock
		handler *handlers.CreatePost
		now     = time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
		input   contract.PostView
	)

	BeforeEach(func() {
		source = &fakeSource{}
		clk = clock.NewMock(now)
		handler = handlers.NewCreatePost(source, clk)
		input = contract.PostView{
			AuthorID:  uuid.New(),
			Title:     "Title",
			Category:  "it",
			Content:   "Content",
			Thumbnail: "https://example.com/t.png",
			Published: now.Add(time.Hour),
		}
	})

	It("passes all fields to the source", func() {
		_, err := handler.Handle(context.Background(), contract.CreatePost{Post: input})
		Expect(err).NotTo(HaveOccurred())

		Expect(source.created.AuthorID).To(Equal(input.AuthorID))
		Expect(source.created.Title).To(Equal(input.Title))
		Expect(source.created.Category).To(Equal(input.Category))
		Expect(source.created.Content).To(Equal(input.Content))
		Expect(source.created.Thumbnail).To(Equal(input.Thumbnail))
		Expect(source.created.Published).To(Equal(input.Published))
	})

	It("sets CreatedAt from the clock, ignoring the input", func() {
		input.CreatedAt = now.Add(-24 * time.Hour)

		_, err := handler.Handle(context.Background(), contract.CreatePost{Post: input})
		Expect(err).NotTo(HaveOccurred())
		Expect(source.created.CreatedAt).To(Equal(now))
	})

	It("returns the post saved by the source", func() {
		source.result = domain.Post{
			ID:        uuid.New(),
			AuthorID:  input.AuthorID,
			Title:     "Saved",
			Category:  "it",
			CreatedAt: now,
		}

		view, err := handler.Handle(context.Background(), contract.CreatePost{Post: input})
		Expect(err).NotTo(HaveOccurred())
		Expect(view.ID).To(Equal(source.result.ID))
		Expect(view.Title).To(Equal("Saved"))
		Expect(view.Category).To(Equal("it"))
		Expect(view.CreatedAt).To(Equal(now))
	})

	It("returns the source error", func() {
		source.err = errors.New("boom")

		view, err := handler.Handle(context.Background(), contract.CreatePost{Post: input})
		Expect(err).To(MatchError("boom"))
		Expect(view).To(BeZero())
	})
})
