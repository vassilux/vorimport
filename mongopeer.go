package main

import (
	"errors"
	"fmt"
	log "github.com/cihub/seelog"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

//
func processPeerMonthlySummary(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var peer = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "monthlypeeroutgoing_summary"
		peer = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "monthlypeerincomming_summary"
		peer = cdr.Peer
	} else {
		return errors.New("Can't detect the call context")
	}
	//
	var id = fmt.Sprintf("%04d%02d-%s", cdr.Calldate.Year(),
		cdr.Calldate.Month(), peer)
	//
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(),
		1, 1, 0, 0, 0, time.UTC)
	//
	var collection = session.DB(config.MongoDbName).C(collectionName)
	metaDoc := PeerMetaData{Peer: peer, Dt: metaDate}

	doc := PeerSummaryCall{Id: id, Meta: metaDoc, Calls: 0, Missing: 0, Duration: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	//
	missing := 0
	if cdr.Disposition > 16 {
		missing = 1
	}
	var change = mgo.Change{
		Update:    bson.M{"$inc": bson.M{"calls": 1, "missing": missing, "duration": cdr.Billsec}},
		ReturnNew: false,
	}
	//
	var info = new(mgo.ChangeInfo)
	info, err = collection.Find(selector).Apply(change, &doc)
	//check if the can execute changes
	if info == nil || info.Updated == 0 {
		err = collection.Insert(doc)
		if err != nil {
			log.Errorf("[peer] Monthly summary insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)

		if info != nil {
			log.Debugf("Monthly new record inserted : %s.", doc.Id)
		} else {
			log.Errorf("Monthly can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		}

	}

	if err != nil {
		log.Debugf("Monthly document [%s] was updated, the update numbers: [%s].\n", doc.Id, info.Updated)
		return err
	}

	return nil
}

//
func processMonthlyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "monthlypeer_outgoing"
		dst = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "monthlypeer_incomming"
		dst = cdr.Peer
	} else {
		return errors.New("Can't detect the call context")
	}
	//
	var id = fmt.Sprintf("%04d%02d-%s-%d", cdr.Calldate.Year(),
		cdr.Calldate.Month(), dst, cdr.Disposition)
	//
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(),
		1, 1, 0, 0, 0, time.UTC)
	//

	var collection = session.DB(config.MongoDbName).C(collectionName)

	metaDoc := MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}

	doc := MonthlyCall{Id: id, Meta: metaDoc, AnswereWaitTime: 0,
		CallMonthly: 0, DurationMonthly: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	var callsDailyInc = fmt.Sprintf("calls_per_days.%d", cdr.Calldate.Day())
	var durationsDailyInc = fmt.Sprintf("durations_per_days.%d", cdr.Calldate.Day())
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"calls": 1, "duration": cdr.Billsec,
			"answer_wait_time": cdr.AnswerWaitTime, callsDailyInc: 1, durationsDailyInc: cdr.Billsec},
		},
		ReturnNew: false,
	}
	//
	var info = new(mgo.ChangeInfo)
	info, err = collection.Find(selector).Apply(change, &doc)
	//check if the can execute changes
	if info == nil || info.Updated == 0 {
		log.Debugf("Monthly update can't be executed , get the error: [ %v], Try execute insert.", err)
		err = collection.Insert(doc)
		if err != nil {
			log.Error("Monthly insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)

		if err != nil {
			log.Errorf("Monthly analytics can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
			return err
		}

	}

	if err != nil {
		log.Errorf("Monthly analytics can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		return err
	}

	return nil
}

func processDailyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "dailypeer_outgoing"
		dst = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "dailypeer_incomming"
		dst = cdr.Peer
	} else {
		return errors.New("Daily analytics can't detect the call context")
	}
	//
	var id = fmt.Sprintf("%04d%02d%02d-%s-%d", cdr.Calldate.Year(), cdr.Calldate.Month(),
		cdr.Calldate.Day(), dst, cdr.Disposition)

	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(), cdr.Calldate.Day(), 1, 0, 0, 0, time.UTC)

	var collection = session.DB(config.MongoDbName).C(collectionName)

	metaDoc := MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}

	doc := DailyCall{Id: id, Meta: metaDoc, AnswereWaitTime: 0, CallDaily: 0,
		DurationDaily: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	var hourlyInc = fmt.Sprintf("calls_per_hours.%d", cdr.Calldate.Hour())
	var durationHourlyInc = fmt.Sprintf("durations_per_hours.%d", cdr.Calldate.Hour())
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"calls": 1, "duration": cdr.Billsec,
			"answer_wait_time": cdr.AnswerWaitTime, hourlyInc: 1, durationHourlyInc: cdr.Billsec},
		},
		ReturnNew: false,
	}
	//
	var info = new(mgo.ChangeInfo)
	info, err = collection.Find(selector).Apply(change, &doc)
	//check if the can execute changes
	if info == nil || info.Updated == 0 {
		log.Debugf("Daily analytics  update can't be executed , get the error: [ %v], Try execute insert.", err)
		err = collection.Insert(doc)
		if err != nil {
			log.Error("Daily analytics  insert failed with error : [%v].", err)
			return err
		}

		info, err = collection.Find(selector).Apply(change, &doc)
		if err != nil {
			log.Debugf("Daily analytics  can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
			return err
		}
	}
	//
	log.Debugf("Document updated : %s\n", doc.Id)
	return nil
}
