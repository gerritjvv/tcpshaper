package bandwidth

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type alwaysErrorLimiter struct {
	Limiter
}

func (l *alwaysErrorLimiter) WaitN(_ context.Context, _ int) error {
	return fmt.Errorf("test error")
}

func (l *alwaysErrorLimiter) Child(conf *RateConfig) Limiter {
	return newBandwidthLimiter(l, conf)
}

func TestLimiter_ParentError(t *testing.T) {
	ctx := context.Background()

	serverConf := NewRateConfig(10, 20)
	connConf := NewRateConfig(5, 10)

	limiter := alwaysErrorLimiter{Limiter: NewBandwidthLimiter(serverConf)}

	err := limiter.Child(connConf).WaitN(ctx, 1)

	if err == nil {
		t.Fatal("expected an error here")
	}

}

func TestLimiter_ChildParentWaitNTimings(t *testing.T) {

	ctx := context.Background()

	serverConf := NewRateConfig(10, 20)
	connConf := NewRateConfig(5, 10)

	limiter := NewBandwidthLimiter(serverConf)

	childLimiters := []Limiter{
		limiter.Child(connConf),
		limiter.Child(connConf),
	}

	var wg sync.WaitGroup
	waitTimesCh := make(chan []time.Duration, len(childLimiters))

	for a, l := range childLimiters {
		wg.Add(1)
		childLimiter := l
		routineIndex := a

		go func() {

			var waitTimes []time.Duration
			startTime := time.Now()

			for i := 0; i < 3; i++ {

				err := childLimiter.WaitN(ctx, 10)
				if err != nil {
					t.Fatalf("no error expected here %s", err)
				}

				afterWaitTime := time.Now()

				timeDiff := afterWaitTime.Sub(startTime)

				// useful debug output for tests
				fmt.Printf("[%d] %d === time : %f\n", routineIndex, i, timeDiff.Seconds())

				// time differences are flaky always, here I try my best.
				if timeDiff.Round(time.Second).Seconds() <= 1 {
					waitTimes = append(waitTimes, 0*time.Second)
				} else {
					waitTimes = append(waitTimes, timeDiff.Round(time.Second))
				}

				startTime = afterWaitTime
			}

			waitTimesCh <- waitTimes
			wg.Done()
		}()

	}

	// note: waitTimeCh is buffered to the exact amount of writes it will receive
	//       otherwise wg.Wait will block forever
	wg.Wait()
	close(waitTimesCh)

	var waitTimes []time.Duration

	for durations := range waitTimesCh {

		for _, d := range durations {
			waitTimes = append(waitTimes, d)
		}
	}

	// From the above config we always expect the following timings:
	// All other timings should be a multiple of two or zero

	var zeroTimingCount = 0
	for _, d := range waitTimes {
		seconds := int(d.Round(time.Second).Seconds())

		if seconds == 0 {
			zeroTimingCount++
		}

		if seconds > 0 && (seconds%2) != 0 {
			t.Fatalf("expected a wait time of a multiple of 2 but got %d", seconds)
		}
	}

	if zeroTimingCount == 0 || zeroTimingCount > len(childLimiters) {
		t.Fatalf("zero or too many zero counts were read %d", zeroTimingCount)
	}
}
