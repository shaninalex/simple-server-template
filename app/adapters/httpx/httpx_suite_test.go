package httpx_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"simple-server-template/app/adapters/httpx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHTTPX(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTPX Suite")
}

var discard = slog.New(slog.DiscardHandler)

func listen(name string, h http.Handler) *httpx.Server {
	s, err := httpx.Listen(name, "127.0.0.1:0", h)
	Expect(err).NotTo(HaveOccurred())
	return s
}

func ok(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") }

// run starts Serve in the background and returns its result channel.
func run(ctx context.Context, timeout time.Duration, servers ...*httpx.Server) <-chan error {
	done := make(chan error, 1)
	go func() { done <- httpx.Serve(ctx, discard, timeout, servers...) }()
	return done
}

var _ = Describe("Listen", func() {
	It("fails right away when the port is taken", func() {
		taken := listen("first", http.HandlerFunc(ok))
		DeferCleanup(taken.Close)

		_, err := httpx.Listen("second", taken.Addr(), http.HandlerFunc(ok))
		Expect(err).To(MatchError(ContainSubstring("second server")))
	})
})

var _ = Describe("Serve", func() {
	It("serves every server until the context is cancelled", func() {
		a, b := listen("a", http.HandlerFunc(ok)), listen("b", http.HandlerFunc(ok))
		ctx, cancel := context.WithCancel(context.Background())
		done := run(ctx, time.Second, a, b)

		for _, s := range []*httpx.Server{a, b} {
			res, err := http.Get("http://" + s.Addr())
			Expect(err).NotTo(HaveOccurred())
			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			_ = res.Body.Close()
		}

		cancel()
		Eventually(done).Should(Receive(BeNil()))

		_, err := net.Dial("tcp", a.Addr())
		Expect(err).To(HaveOccurred(), "listener must be closed after shutdown")
	})

	It("lets in-flight requests finish", func() {
		started, release := make(chan struct{}), make(chan struct{})
		s := listen("slow", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-release
			ok(w, r)
		}))
		ctx, cancel := context.WithCancel(context.Background())
		done := run(ctx, 5*time.Second, s)

		response := make(chan *http.Response, 1)
		go func() {
			defer GinkgoRecover()
			res, err := http.Get("http://" + s.Addr())
			Expect(err).NotTo(HaveOccurred())
			response <- res
		}()

		<-started
		cancel()
		Consistently(done, 100*time.Millisecond).ShouldNot(Receive(), "must wait for the request")

		close(release)
		var res *http.Response
		Eventually(response).Should(Receive(&res))
		Expect(res).To(HaveHTTPStatus(http.StatusOK))
		_ = res.Body.Close()
		Eventually(done).Should(Receive(BeNil()))
	})

	It("gives up on requests that outlive the shutdown timeout", func() {
		started := make(chan struct{})
		s := listen("stuck", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(started)
			<-r.Context().Done()
		}))
		ctx, cancel := context.WithCancel(context.Background())
		done := run(ctx, 50*time.Millisecond, s)

		go func() { _, _ = http.Get("http://" + s.Addr()) }()
		<-started
		cancel()

		var err error
		Eventually(done).Should(Receive(&err))
		Expect(err).To(MatchError(context.DeadlineExceeded))
		_ = s.Close()
	})

	It("stops all servers when one of them fails", func() {
		healthy := listen("healthy", http.HandlerFunc(ok))
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())
		broken := httpx.NewServer("broken", ln, http.HandlerFunc(ok))
		done := run(context.Background(), time.Second, healthy, broken)

		// a listener dying under a running server is a real failure,
		// unlike Server.Close/Shutdown which is a normal stop
		Eventually(func() error {
			_, err := http.Get("http://" + broken.Addr())
			return err
		}).Should(Succeed())
		Expect(ln.Close()).To(Succeed())

		Eventually(done).Should(Receive(&err))
		Expect(err).To(MatchError(ContainSubstring("broken server")))

		_, err = net.Dial("tcp", healthy.Addr())
		Expect(err).To(HaveOccurred(), "healthy server must be stopped too")
	})
})
