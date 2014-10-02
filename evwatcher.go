package main

import (
	"fmt"
	"labix.org/v2/mgo/bson"
	"log"
	"math/big"
)

const (
	EV_NULL          = 0
	EV_START         = 1
	EV_STOP          = 2
	EV_MYSQL_ERROR   = 3
	EV_MYSQL_SUCCESS = 4
	EV_MONGO_ERROR   = 5
	EV_MONGO_SUCCESS = 6
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
	Name  string  // Name of event
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
	ew.event = make(chan *Event)
	ew.done = make(chan bool, 1)

	ew.eventsMask = EmptyBitSet()
	ew.eventsMask.Set(EV_START)
	ew.eventsMask.Set(EV_STOP)
	ew.eventsMask.Set(EV_MYSQL_ERROR)
	ew.eventsMask.Set(EV_MYSQL_SUCCESS)
	ew.eventsMask.Set(EV_MONGO_ERROR)
	ew.eventsMask.Set(EV_MONGO_SUCCESS)

	ew.storageWorker = NewEventStorageWorker()
	//comment cause Save methode used from EventStorageWorker
	//ew.storage = make(chan bson.M)
	//go ew.storageWorker.Work(ew.storage)
	return ew
}

func (eventWatcher *EventWatcher) publishEvent(event bson.M) {
	log.Println("publishEvent : ", event)
	eventWatcher.storageWorker.Save(config.EventsMongoHost, event)
}

//Handler dispatch events and flip the event type
//Flip is used to send one time the same type of notification
func (eventWatcher *EventWatcher) processEvent(event *Event) {
	//log.Println("processEvent : ", event)
	if event.Mask.HasBit(EV_START) {
		var pushEvent = bson.M{"type": EV_START, "data": event.Datas}
		eventWatcher.publishEvent(pushEvent)
	}

	if event.Mask.HasBit(EV_STOP) {
		var pushEvent = bson.M{"type": EV_STOP, "data": event.Datas}
		eventWatcher.publishEvent(pushEvent)
		eventWatcher.done <- true
	}

	//Follow code can/has be refactored
	//mysql parts
	if event.Mask.HasBit(EV_MYSQL_ERROR) {
		if eventWatcher.eventsMask.HasBit(EV_MYSQL_ERROR) {
			var pushEvent = bson.M{"type": EV_MYSQL_ERROR, "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(EV_MYSQL_ERROR)
			eventWatcher.eventsMask.Set(EV_MYSQL_SUCCESS)
		}

	}

	if event.Mask.HasBit(EV_MYSQL_SUCCESS) {
		if eventWatcher.eventsMask.HasBit(EV_MYSQL_SUCCESS) {
			var pushEvent = bson.M{"type": EV_MYSQL_SUCCESS, "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(EV_MYSQL_SUCCESS)
			eventWatcher.eventsMask.Set(EV_MYSQL_ERROR)
		}

	}

	//mongo parts

	if event.Mask.HasBit(EV_MONGO_ERROR) {
		if eventWatcher.eventsMask.HasBit(EV_MONGO_ERROR) {
			var pushEvent = bson.M{"type": EV_MONGO_ERROR, "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(EV_MONGO_ERROR)
			eventWatcher.eventsMask.Set(EV_MONGO_SUCCESS)
		}

	}

	if event.Mask.HasBit(EV_MONGO_SUCCESS) {
		if eventWatcher.eventsMask.HasBit(EV_MONGO_SUCCESS) {
			var pushEvent = bson.M{"type": EV_MONGO_SUCCESS, "data": event.Datas}
			eventWatcher.publishEvent(pushEvent)
			eventWatcher.eventsMask.Clear(EV_MONGO_SUCCESS)
			eventWatcher.eventsMask.Set(EV_MONGO_ERROR)
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
