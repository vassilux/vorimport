package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

const capacity = 32768

type EventStorageWorker struct {
}

func NewEventStorageWorker() (w *EventStorageWorker) {
	return &EventStorageWorker{}
}

func (w *EventStorageWorker) Work(channel chan bson.M) {
	config := GetConfig()
	if config == nil {
		log.Println("config is null ")
	}
	log.Println("EventsMongoHost : " + config.EventsMongoHost)
	for {
		event := <-channel
		w.Save(config.EventsMongoHost, event)
	}
}

func (w *EventStorageWorker) Save(mongoHost string, event bson.M) {
	session, err := mgo.Dial(mongoHost)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	defer session.Close()
	var collection = session.DB("notifications").C("events")
	collection.Insert(event)
}
