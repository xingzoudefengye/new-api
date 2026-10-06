package service

import "sync"

// channelFailures tracks consecutive server-side (5xx) failures per channel.
// Any successful request resets the count; once the count reaches the
// configured threshold the channel is auto-disabled until it recovers via
// manual enable or the periodic channel test.
var channelFailures = struct {
	mu    sync.Mutex
	count map[int]int
}{count: make(map[int]int)}

func recordChannelFailure(channelId int) int {
	channelFailures.mu.Lock()
	defer channelFailures.mu.Unlock()
	channelFailures.count[channelId]++
	return channelFailures.count[channelId]
}

func resetChannelFailure(channelId int) {
	channelFailures.mu.Lock()
	defer channelFailures.mu.Unlock()
	delete(channelFailures.count, channelId)
}
