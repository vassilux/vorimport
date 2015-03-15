package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

//
func isDid(session *mgo.Session, value string) (err error) {
	collection := session.DB(config.MongoDbName).C("dids")
	var selector = bson.M{"did": value}
	//
	var did Did
	err = collection.Find(selector).One(&did)
	//
	if err == nil && did.Did == value {
		return nil
	}
	return err
}

//
func processDidImport(session *mgo.Session, cdr RawCall) (err error) {
	log.Debugf("Import by dnid : %s\n", cdr.Dnid)

	if cdr.Did == "" {
		return
	}

	err = processDidDailyAnalytics(session, cdr)
	if err != nil {
		return err
	}
	err = processDidMonthlyAnalytics(session, cdr)
	if err != nil {
		return err
	}
	err = processDidMonthlyAnalyticsSummary(session, cdr)
	if err != nil {
		return err
	}
	return nil
}

//
//To import data for monthly did summary, just the state of answered and non answered calls
//with calls duration
func processDidMonthlyAnalyticsSummary(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var did = cdr.Did
	collectionName = "monthlydid_summary"
	//
	var id = fmt.Sprintf("%04d%02d-%s", cdr.Calldate.Year(),
		cdr.Calldate.Month(), did)
	//
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(),
		1, 1, 0, 0, 0, time.UTC)
	//
	var collection = session.DB(config.MongoDbName).C(collectionName)
	metaDoc := DidMetaData{Did: did, Dt: metaDate}

	doc := DidSummaryCall{Id: id, Meta: metaDoc, Calls: 0, Missing: 0, Duration: 0}
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
		log.Debugf("Monthly update can't be executed , get the error: [ %v], Try execute insert.", err)
		err = collection.Insert(doc)
		if err != nil {
			log.Errorf("[did] Monthly insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)
		if info != nil {
			log.Debugf("Monthly did new record inserted : %s.", doc.Id)
		} else {
			log.Errorf("Monthly did can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
			return err
		}
	}

	return nil
}

//
func processDidMonthlyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = cdr.Did
	collectionName = "monthlydid_incomming"
	//
	var id = fmt.Sprintf("%04d%02d-%s-%d", cdr.Calldate.Year(),
		cdr.Calldate.Month(), dst, cdr.Disposition)
	//
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(),
		1, 1, 0, 0, 0, time.UTC)
	//

	var collection = session.DB(config.MongoDbName).C(collectionName)
	metaDoc := MetaData{Dst: dst, Disposition: cdr.Disposition, Dt: metaDate}

	doc := MonthlyCall{Id: id, Meta: metaDoc, AnswereWaitTime: 0,
		CallMonthly: 0, DurationMonthly: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	//
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
			log.Errorf("[did] Monthly insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)
		if info != nil {
			log.Debugf("Monthly did new record inserted : %s.", doc.Id)
		} else {
			if err != nil {
				log.Debugf("Document [%s] was updated, the update numbers: [%s].\n", doc.Id, info.Updated)
				return err
			}
		}
	}

	return nil
}

//
func processDidDailyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = cdr.Did
	collectionName = "dailydid_incomming"
	//
	var id = fmt.Sprintf("%04d%02d%02d-%s-%d", cdr.Calldate.Year(), cdr.Calldate.Month(),
		cdr.Calldate.Day(), dst, cdr.Disposition)
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(), cdr.Calldate.Day(), 1, 0, 0, 0, time.UTC)

	var collection = session.DB(config.MongoDbName).C(collectionName)
	metaDoc := MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}

	doc := DailyCall{Id: id, Meta: metaDoc, AnswereWaitTime: 0, CallDaily: 0,
		DurationDaily: 0}

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
		log.Debugf("Daily did update can't be executed , get the error: [ %v], Try execute insert.", err)
		err = collection.Insert(doc)
		if err != nil {
			log.Error("Daily did insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)
		if info != nil {
			log.Debugf("Daily did document updated with success for the document : %s", doc.Id)
			return err
		}

	}
	//
	return nil
}
