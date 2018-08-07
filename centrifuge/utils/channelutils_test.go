package utils

import "testing"

func TestSendNonBlocking(t *testing.T) {
	channel := make(chan interface{})
	SendNonBlocking(true, channel)
	// since this test finishes execution successfully means that the value was sent without blocking
}