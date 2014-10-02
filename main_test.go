package main

import (
	"testing"
	"time"
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

func Test_EventWatcher_MySql(t *testing.T) {
	eventWatcher := NewEventWatcher()
	go eventWatcher.run()
	ev := &Event{
		Mask:  new(BitSet),
		Datas: "EV_MYSQL_ERROR data ev",
		Name:  "EV_MYSQL_ERROR name",
	}

	ev.Mask.Set(EV_MYSQL_ERROR)

	eventWatcher.event <- ev
	//
	ev1 := &Event{
		Mask:  new(BitSet),
		Datas: "EV_MYSQL_ERROR data ev1 ",
		Name:  "EV_MYSQL_ERROR name",
	}
	ev1.Mask.Set(EV_MYSQL_ERROR)
	eventWatcher.event <- ev1

	ev2 := &Event{
		Mask:  new(BitSet),
		Datas: "EV_MYSQL_SUCCESS data ev2",
		Name:  "EV_MYSQL_SUCCESS name",
	}

	ev2.Mask.Set(EV_MYSQL_SUCCESS)
	eventWatcher.event <- ev2

	ev3 := &Event{
		Mask:  new(BitSet),
		Datas: "EV_MYSQL_SUCCESS data ev3",
		Name:  "EV_MYSQL_SUCCESS name",
	}
	ev3.Mask.Set(EV_MYSQL_SUCCESS)
	eventWatcher.event <- ev3

	ev4 := &Event{
		Mask:  new(BitSet),
		Datas: "EV_MYSQL_ERROR data ev4",
		Name:  "EV_MYSQL_ERROR name",
	}
	ev4.Mask.Set(EV_MYSQL_ERROR)
	eventWatcher.event <- ev4
	time.Sleep(1 * time.Second)
	if eventWatcher.eventsMask.HasBit(EV_MYSQL_ERROR) || !eventWatcher.eventsMask.HasBit(EV_MYSQL_SUCCESS) {
		t.Fail()
	}

}
