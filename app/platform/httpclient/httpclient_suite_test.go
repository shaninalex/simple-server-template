package httpclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"simple-server-template/app/core/errs"
	"simple-server-template/app/platform/httpclient"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTTPClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Client Suite")
}

type payload struct {
	Name string `json:"name"`
}

func newClient(url string, retries int) *httpclient.Client {
	c, err := httpclient.New(httpclient.Config{
		BaseURL:      url,
		Timeout:      time.Second,
		Retries:      retries,
		RetryBackoff: time.Millisecond,
	})
	Expect(err).NotTo(HaveOccurred())
	return c
}

func serve(h http.HandlerFunc) *httptest.Server {
	srv := httptest.NewServer(h)
	DeferCleanup(srv.Close)
	return srv
}

var _ = Describe("Client", func() {
	Describe("New", func() {
		It("rejects an empty base URL", func() {
			_, err := httpclient.New(httpclient.Config{})
			Expect(err).To(HaveOccurred())
		})

		It("rejects a non-http scheme", func() {
			_, err := httpclient.New(httpclient.Config{BaseURL: "ftp://example.com"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Post", func() {
		It("sends JSON and decodes the response", func() {
			srv := serve(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				Expect(r.Method).To(Equal(http.MethodPost))
				Expect(r.URL.Path).To(Equal("/items"))
				Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))
				Expect(r.Header.Get("Accept")).To(Equal("application/json"))

				var in payload
				Expect(json.NewDecoder(r.Body).Decode(&in)).To(Succeed())
				Expect(in.Name).To(Equal("sent"))

				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(payload{Name: "created"})
			})

			var out payload
			err := newClient(srv.URL, 0).Post(context.Background(), "/items", payload{Name: "sent"}, &out)
			Expect(err).NotTo(HaveOccurred())
			Expect(out.Name).To(Equal("created"))
		})

		It("does not retry on a server error", func() {
			var calls atomic.Int32
			srv := serve(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusBadGateway)
			})

			err := newClient(srv.URL, 2).Post(context.Background(), "/items", payload{}, &payload{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindPlatform))
			Expect(calls.Load()).To(BeEquivalentTo(1))
		})

		It("does not retry on rate limiting", func() {
			var calls atomic.Int32
			srv := serve(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusTooManyRequests)
			})

			err := newClient(srv.URL, 2).Post(context.Background(), "/items", payload{}, &payload{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindRateLimited))
			Expect(calls.Load()).To(BeEquivalentTo(1))
		})

		It("maps 404 to a not-found error", func() {
			srv := serve(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			})

			err := newClient(srv.URL, 0).Post(context.Background(), "/items", payload{}, &payload{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindNotFound))
		})

		It("fails on a payload that cannot be encoded", func() {
			err := newClient("http://localhost", 0).Post(context.Background(), "/items", make(chan int), &payload{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindPlatform))
		})
	})

	Describe("GetJSON", func() {
		It("passes query params and decodes the response", func() {
			srv := serve(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				Expect(r.Method).To(Equal(http.MethodGet))
				Expect(r.URL.Query().Get("_limit")).To(Equal("5"))
				_ = json.NewEncoder(w).Encode([]payload{{Name: "a"}})
			})

			var out []payload
			err := newClient(srv.URL, 0).GetJSON(context.Background(), "/items", map[string][]string{"_limit": {"5"}}, &out)
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(Equal([]payload{{Name: "a"}}))
		})

		It("retries a server error and gives up after the configured retries", func() {
			var calls atomic.Int32
			srv := serve(func(w http.ResponseWriter, _ *http.Request) {
				calls.Add(1)
				w.WriteHeader(http.StatusBadGateway)
			})

			err := newClient(srv.URL, 2).GetJSON(context.Background(), "/items", nil, &payload{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindPlatform))
			Expect(calls.Load()).To(BeEquivalentTo(3))
		})

		It("maps 429 to a rate-limited error", func() {
			srv := serve(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusTooManyRequests)
			})

			err := newClient(srv.URL, 0).GetJSON(context.Background(), "/items", nil, &payload{})
			Expect(errs.KindOf(err)).To(Equal(errs.KindRateLimited))
		})
	})
})
