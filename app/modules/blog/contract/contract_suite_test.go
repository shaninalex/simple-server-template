package contract_test

import (
	"simple-server-template/app/modules/blog/contract"
	"testing"
	"time"
	"uuid"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestContract(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Contract Suite")
}

var _ = Describe("Posts", func() {
	var testPost *contract.PostView
	published := time.Now()
	createdDate := published.Add(-time.Hour * 2)

	BeforeEach(func() {
		testPost = &contract.PostView{
			ID:        uuid.New(),
			Title:     "Les Miserables",
			Content:   "this is content",
			Category:  "test",
			AuthorID:  uuid.New(),
			Published: published,
			CreatedAt: createdDate,
		}
	})

	Describe("Post", func() {
		Context("published date", func() {
			It("before created date", func() {
				Expect(testPost.CreatedAt).Should(BeTemporally("<", testPost.Published))
			})
			It("might be a zero time", func() {
				testPost.Published = time.Time{}
				Expect(testPost.Published.IsZero()).Should(BeTrue())
			})
		})

		Context("ID", func() {
			It("Have to be valid uuid", func() {
				id, err := uuid.Parse(testPost.ID.String())
				Expect(err == nil).Should(BeTrue())
				Expect(id).Should(Equal(testPost.ID))
			})
		})
	})
})
