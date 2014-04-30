package main

import (
	"errors"
	"fmt"
	log "github.com/cihub/seelog"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"time"
	m "vorimport/models"
)

const (
	ODIN_MONGO_DATABASENAME = "revor"
)

//
func createMongoCdr(session *mgo.Session, cdr m.RawCall) (err error) {
	collection := session.DB(ODIN_MONGO_DATABASENAME).C("cdrs")
	//
	err = collection.Insert(cdr)
	if err != nil {
		log.Criticalf("Can't insert document: %v", err)
		os.Exit(1)
	} else {
		log.Debugf("Row inserted into mongo database: %s", cdr.ClidName)
	}
	return
}

//
func processMonthlyAnalytics(session *mgo.Session, cdr m.RawCall) (err error) {
	//
	var collectionName = ""
	var dst = ""
	if cdr.InoutStatus == DIRECTION_CALL_OUT {
		collectionName = "monthlyanalytics_outgoing"
		dst = cdr.Src
	} else if cdr.InoutStatus == DIRECTION_CALL_IN {
		collectionName = "monthlyanalytics__incomming"
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
	metaDoc := m.MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}
	doc := m.MonthlyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime,
		CallMonthly: 0, DurationMonthly: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"call_monthly": 1, "duration_monthly": cdr.Billsec,
			"answer_wait_time": cdr.AnswerWaitTime},
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
			log.Debugf("Monthly analytics wew record inserted : %s.", doc.Id)
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

func processDailyAnalytics(session *mgo.Session, cdr m.RawCall) (err error) {
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
	//var t = time.Unix(cdr.calldate, 0)
	var id = fmt.Sprintf("%04d%02d%02d-%s-%d", cdr.Calldate.Year(), cdr.Calldate.Month(),
		cdr.Calldate.Day(), dst, cdr.Disposition)
	var metaDate = time.Date(cdr.Calldate.Year(), cdr.Calldate.Month(), cdr.Calldate.Day(), 1, 0, 0, 0, time.UTC)
	log.Debugf("Import daily analytics :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := m.MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}
	doc := m.DailyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime, CallDaily: 0,
		DurationDaily: 0}
	//err = collection.Insert(doc)
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	var hourlyInc = fmt.Sprintf("call_hourly.%d", cdr.Calldate.Hour())
	var durationHourlyInc = fmt.Sprintf("duration_hourly.%d", cdr.Calldate.Hour())
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"call_daily": 1, "duration_daily": cdr.Billsec,
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
			log.Debugf("Daily document updated with success for the document : %s", doc.Id)
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
func processDidImport(session *mgo.Session, cdr m.RawCall) (err error) {
	log.Debugf("Import by did : %s\n", cdr.Dnid)
	err = processDidDailyAnalytics(session, cdr)
	if err != nil {
		return nil
	}
	err = processDidMonthlyAnalytics(session, cdr)
	return nil
}

//
func processDidMonthlyAnalytics(session *mgo.Session, cdr m.RawCall) (err error) {
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
	metaDoc := m.MetaData{Dst: dst, Disposition: cdr.Disposition, Dt: metaDate}
	doc := m.MonthlyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime,
		CallMonthly: 0, DurationMonthly: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"call_monthly": 1, "duration_monthly": cdr.Billsec,
			"answer_wait_time": cdr.AnswerWaitTime},
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
func processDidDailyAnalytics(session *mgo.Session, cdr m.RawCall) (err error) {
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
	metaDoc := m.MetaData{Dst: dst, Dt: metaDate, Disposition: cdr.Disposition}
	doc := m.DailyCall{Id: id, Meta: metaDoc, AnswereWaitTime: cdr.AnswerWaitTime, CallDaily: 0,
		DurationDaily: 0}
	//err = collection.Insert(doc)
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	var hourlyInc = fmt.Sprintf("call_hourly.%d", cdr.Calldate.Hour())
	var durationHourlyInc = fmt.Sprintf("duration_hourly.%d", cdr.Calldate.Hour())
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"call_daily": 1, "duration_daily": cdr.Billsec,
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

func importCdrToMongo(session *mgo.Session, cdr m.RawCall) (err error) {
	log.Debugf("Start analyze data for mongo database.")
	createMongoCdr(session, cdr)
	err = processDailyAnalytics(session, cdr)
	if err != nil {
		return err
	}
	err = processMonthlyAnalytics(session, cdr)
	if err != nil {
		return err
	}
	if cdr.InoutStatus == DIRECTION_CALL_IN {
		err = processDidImport(session, cdr)
	}
	err = nil
	return
}
