package main

import (
	"errors"
	"fmt"
	log "github.com/cihub/seelog"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"time"
)

type TestDummy struct {
	FirstName string
}

const (
	ODIN_MONGO_DATABASENAME = "revor"
)

//
func createMongoCdr(session *mgo.Session, cdr RawCall) (err error) {
	collection := session.DB(ODIN_MONGO_DATABASENAME).C("cdrs")
	//
	err = collection.Insert(cdr)
	if err != nil {
		log.Criticalf("Can't insert document: %v", err)
		os.Exit(1)
	} else {
		log.Debugf("Row inserted into mongo database for %s from asterisk [%s]", cdr.ClidName, cdr.AsteriskId)
	}
	return
}

//
//
//To import data for monthly did summary, just the state of answered and non answered calls
//with calls duration
func processPeerMonthlySummary(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var peer = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "monthlypeeroutgoing_summary"
		peer = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "monthlypeerincomming_summary"
		peer = cdr.Dst
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
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
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
			log.Debugf("[peer] Monthly new record inserted : %s.", doc.Id)
		} else {
			log.Errorf("[peer] Monthly can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		}
	} else {
		if err != nil {
			log.Debugf("[peer] Monthly document [%s] was updated, the update numbers: [%s].\n", doc.Id, info.Updated)
		} else {
			return err
		}

	}
	//
	return nil
}

//
func processMonthlyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "monthlyanalytics_outgoing"
		dst = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "monthlyanalytics_incomming"
		dst = cdr.Dst
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
	log.Debugf("Import monthly analytics :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}
	doc := MonthlyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime,
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
		if info != nil {
			log.Tracef("Monthly analytics wew record inserted : %s.", doc.Id)
		} else {
			log.Errorf("Monthly analytics can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		}
	} else {
		if err != nil {
			log.Debugf("Document [%s] was updated, the update numbers: [%s].", doc.Id, info.Updated)
		} else {
			return err
		}

	}

	//
	return nil
}

func processDailyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "dailyanalytics_outgoing"
		dst = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "dailyanalytics_incomming"
		dst = cdr.Dst
	} else {
		return errors.New("[mongo] Can't detect the call context")
	}
	//
	var id = fmt.Sprintf("%04d%02d%02d-%s-%d", cdr.Calldate.Year(), cdr.Calldate.Month(),
		cdr.Calldate.Day(), dst, cdr.Disposition)
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(), cdr.Calldate.Day(), 1, 0, 0, 0, time.UTC)
	log.Debugf("Import daily analytics :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}
	doc := DailyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime, CallDaily: 0,
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
		log.Debugf("Daily update can't be executed , get the error: [ %v], Try execute insert.", err)
		err = collection.Insert(doc)
		if err != nil {
			log.Error("Daily insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)
		if info != nil {
			log.Debugf("Daily docCallsHourlyument updated with success for the document : %s", doc.Id)
		} else {
			log.Debugf("Daily document can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		}
	} else {
		log.Debugf("Document updated : %s\n", doc.Id)
	}
	//
	return nil
}

//
func processDidImport(session *mgo.Session, cdr RawCall) (err error) {
	log.Tracef("Import by did : %s\n", cdr.Dnid)
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
	var did = cdr.Dnid
	collectionName = "monthlydid_summary"
	//
	var id = fmt.Sprintf("%04d%02d-%s", cdr.Calldate.Year(),
		cdr.Calldate.Month(), did)
	//
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(),
		1, 1, 0, 0, 0, time.UTC)
	//
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
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
		}
	} else {
		if err != nil {
			log.Debugf("Document [%s] was updated, the update numbers: [%s].\n", doc.Id, info.Updated)
		} else {
			return err
		}

	}
	//
	return nil
}

//
func processDidMonthlyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = cdr.Dnid
	collectionName = "monthlydid_incomming"
	//
	var id = fmt.Sprintf("%04d%02d-%s-%d", cdr.Calldate.Year(),
		cdr.Calldate.Month(), dst, cdr.Disposition)
	//
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(),
		1, 1, 0, 0, 0, time.UTC)
	//
	log.Debugf("Import monthly did :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := MetaData{Dst: dst, Disposition: cdr.Disposition, Dt: metaDate}
	doc := MonthlyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime,
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
			log.Errorf("Monthly did can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		}
	} else {
		if err != nil {
			log.Debugf("Document [%s] was updated, the update numbers: [%s].\n", doc.Id, info.Updated)
		} else {
			return err
		}

	}
	//
	return nil
}

//
func processDidDailyAnalytics(session *mgo.Session, cdr RawCall) (err error) {
	//
	var collectionName = ""
	var dst = cdr.Dnid
	collectionName = "dailydid_incomming"
	//
	var id = fmt.Sprintf("%04d%02d%02d-%s-%d", cdr.Calldate.Year(), cdr.Calldate.Month(),
		cdr.Calldate.Day(), dst, cdr.Disposition)
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(), cdr.Calldate.Day(), 1, 0, 0, 0, time.UTC)
	log.Debugf("Import daily did :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}

	doc := DailyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime, CallDaily: 0,
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
		} else {
			log.Debugf("Daily did document can't be updated, get the error : [%v] for the document : %s", err, doc.Id)
		}
	} else {
		log.Debugf("Document did updated : %s\n", doc.Id)
	}
	//
	return nil
}

func importCdrToMongo(session *mgo.Session, cdr RawCall) (err error) {
	log.Debugf("Start analyze data for mongo database from asterisk : [%s].", cdr.AsteriskId)
	createMongoCdr(session, cdr)
	err = processDailyAnalytics(session, cdr)
	if err != nil {
		return err
	}
	err = processMonthlyAnalytics(session, cdr)
	if err != nil {
		return err
	}

	err = processPeerMonthlySummary(session, cdr)
	if err != nil {
		return err

	}

	if cdr.InoutStatus == DIRECTION_CALL_IN {
		err = processDidImport(session, cdr)
	}
	err = nil
	return
}
