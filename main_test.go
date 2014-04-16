package main

import (
	"testing"
)

func Test_DstChannel(t *testing.T) {
	var channel = "SIP/6006-01010101"
	peer := getPeerFromChannel(channel)
	if peer != "6006" {
		t.Error("It is not good peer for channel [%s].", channel)
	}
	channel = "DAHDI/g1/0493948400-01010101"
	peer = getPeerFromChannel(channel)
	if peer != "0493948400" {
		t.Error("It is not good peer [%s] for channel [%s].", peer, channel)
	}
	t.Log("dstChannelTester test passed.")
}
