package service

import "testing"

func TestChannelFailureCounterCountsAndResets(t *testing.T) {
	const channel = 990001
	defer resetChannelFailure(channel)

	if got := recordChannelFailure(channel); got != 1 {
		t.Fatalf("first failure count = %d, want 1", got)
	}
	if got := recordChannelFailure(channel); got != 2 {
		t.Fatalf("second failure count = %d, want 2", got)
	}
	resetChannelFailure(channel)
	if got := recordChannelFailure(channel); got != 1 {
		t.Fatalf("count after reset = %d, want 1", got)
	}
}

func TestChannelFailureCounterIsolatedPerChannel(t *testing.T) {
	defer func() {
		resetChannelFailure(990002)
		resetChannelFailure(990003)
	}()

	recordChannelFailure(990002)
	recordChannelFailure(990002)
	if got := recordChannelFailure(990003); got != 1 {
		t.Fatalf("unrelated channel count = %d, want 1", got)
	}
}
