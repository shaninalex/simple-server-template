package clock_test

import (
	"sync"
	"testing"
	"time"

	"simple-server-template/app/platform/clock"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestClock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Clock Suite")
}

var _ = Describe("System", func() {
	It("returns the current time", func() {
		before := time.Now()
		now := clock.System().Now()
		after := time.Now()

		Expect(now).To(BeTemporally(">=", before))
		Expect(now).To(BeTemporally("<=", after))
	})
})

var _ = Describe("Mock", func() {
	var (
		initial = time.Date(2026, time.September, 12, 12, 0, 0, 0, time.UTC)
		mock    *clock.Mock
	)

	BeforeEach(func() {
		mock = clock.NewMock(initial)
	})

	It("starts at the given time", func() {
		Expect(mock.Now()).To(Equal(initial))
	})

	It("does not move by itself", func() {
		Consistently(mock.Now).WithTimeout(20 * time.Millisecond).Should(Equal(initial))
	})

	It("advances by the given duration", func() {
		mock.Advance(15 * time.Minute)
		Expect(mock.Now()).To(Equal(initial.Add(15 * time.Minute)))

		mock.Advance(time.Hour)
		Expect(mock.Now()).To(Equal(initial.Add(75 * time.Minute)))
	})

	It("can be set to an arbitrary time", func() {
		next := time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC)
		mock.Set(next)
		Expect(mock.Now()).To(Equal(next))
	})

	It("is safe for concurrent use", func() {
		var wg sync.WaitGroup
		for range 50 {
			wg.Go(func() { mock.Advance(time.Second) })
			wg.Go(func() { _ = mock.Now() })
		}
		wg.Wait()

		Expect(mock.Now()).To(Equal(initial.Add(50 * time.Second)))
	})

	Describe("Frozen", func() {
		It("behaves like a mock started at the given time", func() {
			Expect(clock.Frozen(initial).Now()).To(Equal(initial))
		})
	})
})
