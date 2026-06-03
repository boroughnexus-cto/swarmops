package main

import (
	"context"
	"testing"
	"time"
)

func TestNextBackoff(t *testing.T) {
	const max = 30 * time.Second
	cases := []struct {
		cur, want time.Duration
	}{
		{time.Second, 2 * time.Second},
		{2 * time.Second, 4 * time.Second},
		{16 * time.Second, 30 * time.Second}, // 32s would exceed max → capped
		{30 * time.Second, 30 * time.Second}, // already at cap → stays
	}
	for _, c := range cases {
		if got := nextBackoff(c.cur, max); got != c.want {
			t.Errorf("nextBackoff(%s, %s) = %s, want %s", c.cur, max, got, c.want)
		}
	}
}

func TestSleepCtxCompletes(t *testing.T) {
	if !sleepCtx(context.Background(), 5*time.Millisecond) {
		t.Fatal("sleepCtx returned false for a timer that should elapse")
	}
}

func TestSleepCtxCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before sleeping
	if sleepCtx(ctx, time.Hour) {
		t.Fatal("sleepCtx returned true despite cancelled context")
	}
}
