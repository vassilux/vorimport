package main

import (
	"strings"
	"testing"
	//"time"
)

func getPeerFromChannel(channel string) (peer string) {

	if strings.Contains(channel, "IAX2/trunk_") {
		return ""
	}

	//try to find the destination from channel
	delim := '-'
	if strings.Contains(channel, "@") {
		delim = '@'
	}

	w := strings.FieldsFunc(channel, func(r rune) bool {
		switch r {

		case '/', delim:
			return true
		}
		return false
	})

	if len(w) >= 3 {
		return w[len(w)-2]
	} else {
		return channel
	}

}

func Test_DstChannel(t *testing.T) {
	var channel = "SIP/6006-01010101"
	peer := getPeerFromChannel(channel)
	if peer != "6006" {
		t.Error("It is not good peer for channel [%s].", channel)
		t.Fail()
	}
	channel = "DAHDI/g1/0493948400-01010101"
	peer = getPeerFromChannel(channel)
	if peer != "0493948400" {
		t.Error("It is not good peer [%s] for channel [%s].", peer, channel)
		t.Fail()
	}
	t.Log("dstChannelTester test passed.")
	channel = "'DAHDI/132-1'"
	peer = getPeerFromChannel(channel)
	if peer != "132" {
		t.Error("It is not good peer [%s] for channel [%s].", peer, channel)
		t.Fail()
	}

	channel = "Local/8129@DLPN_DialPlan1-000000e6;1"
	peer = getPeerFromChannel(channel)
	if peer != "8129" {
		t.Errorf("It is not good peer [%s] for channel [%s].", peer, channel)
		t.Fail()
	}

	channel = "IAX2/trunk_2-13959"
	peer = getPeerFromChannel(channel)
	if peer != "" {
		t.Errorf("It is not good peer [%s] for channel [%s].", peer, channel)
		t.Fail()
	}

	t.Log("dstChannelTester test passed.")
}

/*func Test_EventWatcher_MySql(t *testing.T) {
	loadConfig(false)
	config := GetConfig()
	eventWatcher := NewEventWatcher(config)
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

}*/

/*func Test_CallOriginator(t *testing.T) {
	addr := "192.168.3.20"
	port := 5038
	user := "astmanager"
	pswd := "lepanos"

	callOriginator := NewCallOriginator(addr, port, user, pswd)

	callOriginator.testCall <- true

	time.Sleep(1 * time.Second)

}*/
