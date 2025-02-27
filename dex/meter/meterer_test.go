package meter

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDelayedRelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	minDelay := time.Millisecond * 20
	out, in := DelayedRelay(ctx, minDelay, 4)
	defer func() { <-out }() // wait for channel close (relay shutdown)
	defer cancel()

	times := make([]time.Time, 0)

	countSignals := func() (n int) {
		for {
			select {
			case <-out:
				times = append(times, time.Now())
				n++
			case <-time.After(minDelay * 2):
				return
			}
		}
	}

	lastDelay := func() time.Duration {
		return times[len(times)-1].Sub(times[len(times)-2])
	}

	sendSignals := func(n int, err error) {
		for i := 0; i < n; i++ {
			in <- err
		}
	}

	go sendSignals(1, nil)
	if n := countSignals(); n != 1 {
		t.Fatalf("wrong signal count. expected 1, got %d", n)
	}

	// Sending two signals simultaneously should end up with two signals, at
	// least minDelay apart.
	go sendSignals(2, nil)
	if n := countSignals(); n != 2 {
		t.Fatalf("wrong signal count. expected 2, got %d", n)
	}
	if d := lastDelay(); d < minDelay {
		t.Fatalf("last delay was less than minimum. %s < %s", d, minDelay)
	}

	// Should get the same exact result for 3 or more simultaneous signals.
	go sendSignals(5, nil)
	if n := countSignals(); n != 2 {
		t.Fatalf("wrong signal count. expected reduction to 2, got %d", n)
	}
	if d := lastDelay(); d < minDelay {
		t.Fatalf("last delay was less than minimum. %s < %s", d, minDelay)
	}

	// Errors should go right through.
	go sendSignals(10, errors.New(""))
	if n := countSignals(); n != 10 {
		t.Fatalf("wrong signal count. expected 10 errors, got %d", n)
	}
}
