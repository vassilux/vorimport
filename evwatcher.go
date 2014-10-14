package main

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"log"
	"math/big"
)

const (
	EV_NULL = 0
	APPSTA  = 1
	APPSTO  = 2
	MYSQKO  = 3
	MYSQOK  = 4
	MONGOKO = 5
	MONGOOK = 6
	TCALOK  = 7
	TCALKO  = 8
	CCALOK  = 9 // check call action success
	CCALKO  = 10
)

type BitSet struct {
	bits big.Int
}

func (b *BitSet) Set(bit int) *BitSet {
	b.bits.SetBit(&b.bits, int(bit), 1)
	return b
}

func (b *BitSet) Clear(bit int) *BitSet {
	b.bits.SetBit(&b.bits, bit, 0)
	return b
}

func (b *BitSet) HasBit(bit int) bool {
	return b.bits.Bit(bit) == 1
}

func (b *BitSet) Flip(bit int) *BitSet {
	if !b.HasBit(bit) {
		return b.Set(bit)
	}
	return b.Clear(bit)
}

func (b BitSet) String() string {
	return fmt.Sprintf("%b", &b.bits)
}

func EmptyBitSet() *BitSet {
	return new(BitSet)
}

type Event struct {
	Mask  *BitSet // Mask of event
	Datas string  // Data of event
	Code  string  // Code of event
}

type EventWatcher struct {
	event chan *Event // Events are returned on this channel
	done  chan bool   // Channel for sending a "quit message"
	//storage       chan bson.M
	isClosed      bool // Set to true when Close() is first called
	eventsMask    *BitSet
	storageWorker *EventStorageWorker
	config        *Config
}

//Create a event watcher and set the mask of events to all
func NewEventWatcher(config *Config) (ew *EventWatcher) {
	ew = new(EventWatcher)
	ew.config = config
	ew.event = make(chan *Event)
	ew.done = make(chan bool, 1)

	ew.eventsMask = EmptyBitSet()
	ew.eventsMask.Set(APPSTA)
	ew.eventsMask.Set(APPSTO)
	ew.eventsMask.Set(MYSQKO)
	ew.eventsMask.Set(MYSQOK)
	ew.eventsMask.Set(MONGOKO)
	ew.eventsMask.Set(MONGOOK)
	ew.eventsMask.Set(TCALKO)
	ew.eventsMask.Set(TCALOK)
	ew.eventsMask.Set(CCALOK)
	ew.eventsMask.Set(CCALKO)

	ew.storageWorker = NewEventStorageWorker()
	//comment cause Save methode used from EventStorageWorker
	//ew.storage = make(chan bson.M)
	//go ew.storageWorker.Work(ew.storage)
	return ew
}

func (eventWatcher *EventWatcher) publishEvent(event bson.M) {
	for _, notification := range eventWatcher.config.Notifications {
		event["appid"] = "vorimport"
		event["asteriskid"] = eventWatcher.config.AsteriskID
		event["transport"] = notification
		log.Println("publishEvent : ", event)
		eventWatcher.storageWorker.Save(config.EventsMongoHost, event)
	}
}

//Handler dispatch events and flip the event type
//Flip is used to send one time the same type of notification
func (eventWatcher *EventWatcher) processEvent(event *Event) {
	//log.Println("processEvent : ", event)
	if event.Mask.HasBit(APPSTA) {
		var pushEvent = bson.M{"type": 1, "code": "APPSTA", "data": event.Datas}
		eventWatcher.publishEvent(pushEvent)
	}

	if event.Mask.HasBit(APPSTO) {
		var pushEvent = bson.M{"type": 1, "code": "APPSTO", "data": event.Datas}
		eventWatcher.publishEvent(pushEvent)
		eventWatcher.done <- true
	}

	//Follow code can/has be refactored
	//mysql parts
	if event.Mask.HasBit(MYSQKO) {
		if eventWatcher.eventsMask.HasBit(MYSQKO) {
			var pushEvent = bson.M{"type": 1, "code": "MYSQKO", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(MYSQKO)
			eventWatcher.eventsMask.Set(MYSQOK)
		}

	}

	if event.Mask.HasBit(MYSQOK) {
		if eventWatcher.eventsMask.HasBit(MYSQOK) {
			var pushEvent = bson.M{"type": 1, "code": "MYSQOK", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(MYSQOK)
			eventWatcher.eventsMask.Set(MYSQKO)
		}

	}

	//mongo parts

	if event.Mask.HasBit(MONGOKO) {
		if eventWatcher.eventsMask.HasBit(MONGOKO) {
			var pushEvent = bson.M{"type": 1, "code": "MONGOKO", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(MONGOKO)
			eventWatcher.eventsMask.Set(MONGOOK)
		}

	}

	if event.Mask.HasBit(MONGOOK) {
		if eventWatcher.eventsMask.HasBit(MONGOOK) {
			var pushEvent = bson.M{"type": 1, "code": "MONGOOK", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(MONGOOK)
			eventWatcher.eventsMask.Set(MONGOKO)
		}

	}

	//test call part
	if event.Mask.HasBit(TCALKO) {
		if eventWatcher.eventsMask.HasBit(TCALKO) {
			var pushEvent = bson.M{"type": 1, "code": "TCALKO", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(TCALKO)
			eventWatcher.eventsMask.Set(TCALOK)
		}

	}

	if event.Mask.HasBit(TCALOK) {
		if eventWatcher.eventsMask.HasBit(TCALOK) {
			var pushEvent = bson.M{"type": 1, "code": "TCALOK", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(TCALOK)
			eventWatcher.eventsMask.Set(TCALKO)
		}

	}

	if event.Mask.HasBit(CCALKO) {
		if eventWatcher.eventsMask.HasBit(CCALKO) {
			var pushEvent = bson.M{"type": CCALKO, "code": "CCALKO", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(CCALKO)
			eventWatcher.eventsMask.Set(CCALOK)
		}

	}

	if event.Mask.HasBit(CCALOK) {
		fmt.Println("Enter CCALOK")
		if eventWatcher.eventsMask.HasBit(CCALOK) {
			fmt.Println("Enter CCALOK send and change flag")
			var pushEvent = bson.M{"type": 1, "code": "CCALOK", "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(CCALOK)
			eventWatcher.eventsMask.Set(CCALKO)
		}

	}

}

func (eventWatcher *EventWatcher) run() {
	for {
		select {
		case c := <-eventWatcher.event:
			eventWatcher.processEvent(c)
		}
	}
}
