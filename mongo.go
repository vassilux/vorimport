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

const (
	ODIN_MONGO_DATABASENAME = "asterisk"
)

type rawCall struct {
	Id             bson.ObjectId `bson:"_id"`
	Calldate       time.Time     `bson:"calldate"`
	MetadataDt     time.Time     `bson:"metadataDt"`
	ClidName       string        `bson:"clidName"`
	ClidNumber     string        `bson:"clidNumber"`
	Src            string        `bson:"src"`
	Channel        string        `bson:"channel"`
	Dcontext       string        `bson:"dcontext"`
	Disposition    int           `bson:"disposition"`
	Answerwaittime int64         `bson:"answerwaittime"`
	Billsec        int           `bson:"billsec"`
	Duration       int           `bson:"duration"`
	Uniqueid       string        `bson:"uniqueid"`
	Inoutstatus    int           `bson:"inoutstatus"`
	Recordfile     string        `bson:"recordfile"`
	Dst            string        `bson:"dst"`
}

type metaData struct {
	User        string    `bson:"user"`
	Dt          time.Time `bson:"dt"`
	Disposition int       `bson:"disposition"`
}

type dailyCall struct {
	Id              string   `bson:"_id"`
	MetaData        metaData `bson:"metadata"`
	AnswereWaitTime int64    `bson:"answere_wait_time"`
	CallDaily       int      `bson:"call_daily"`
	DurationDaily   int      `bson:"duration_daily"`
}

type monthlyCall struct {
	Id              string   `bson:"_id"`
	MetaData        metaData `bson:"metadata"`
	AnswereWaitTime int64    `bson:"answere_wait_time"`
	CallMonthly     int      `bson:"call_monthly"`
	DurationMonthly int      `bson:"duration_monthly"`
}

func createMongoCdr(session *mgo.Session, cdr Cdr) (err error) {
	collection := session.DB(ODIN_MONGO_DATABASENAME).C("cdrs")
	//
	/*now := time.Now()
	var _, offset = now.Zone()
	var mongoCalldate = cdr.calldate + int64(offset)
	var mongoNow = now.Unix() + int64(offset)*/
	doc := rawCall{Id: bson.NewObjectId(), Calldate: cdr.calldate, MetadataDt: time.Unix(time.Now().Unix()+int64(timeZoneOffset), 0), ClidName: cdr.clid,
		ClidNumber: cdr.clid, Src: cdr.src, Channel: cdr.channel, Dcontext: cdr.dcontext, Disposition: cdr.causeStatus,
		Answerwaittime: cdr.waitAnswer, Billsec: cdr.billsec, Duration: cdr.duration, Uniqueid: cdr.uniqueid, Inoutstatus: cdr.inoutstatus,
		Recordfile: cdr.recordfile, Dst: cdr.dst}
	err = collection.Insert(doc)
	if err != nil {
		log.Criticalf("Can't insert document: %v", err)
		os.Exit(1)
	} else {
		log.Debugf("Row inserted into mongo database: %s", doc.ClidName)
	}
	return
}

/**
{
  "_id": "201306-6000-19",
  "answere_wait_time": NumberInt(0),
  "call_monthly": NumberInt(1),
  "duration_monthly": NumberInt(0),
  "metadata": {
    "user": "6000",
    "dt": ISODate("2013-06-01T01:00:00.0Z"),
    "disposition": NumberInt(19)
  }
}
**/
func processMonthlyAnalytics(session *mgo.Session, cdr Cdr) (err error) {
	//
	var collectionName = ""
	if cdr.inoutstatus == DIRECTION_CALL_OUT {
		collectionName = "monthlyanalytics_outgoing"
	} else if cdr.inoutstatus == DIRECTION_CALL_IN {
		collectionName = "monthlyanalytics__incomming"
	} else {
		return errors.New("Can't detect the call context")
	}
	//
	var id = fmt.Sprintf("%04d%02d-%s-%d", cdr.calldate.Year(), cdr.calldate.Month(), cdr.src, cdr.causeStatus)
	var metaDate = time.Date(cdr.calldate.Year(), cdr.calldate.Month(), cdr.calldate.Day(), 1, 0, 0, 0, time.UTC)
	//
	log.Debugf("Import monthly analytics :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := metaData{User: cdr.src, Dt: metaDate, Disposition: cdr.causeStatus}
	doc := monthlyCall{Id: id, MetaData: metaDoc, AnswereWaitTime: cdr.waitAnswer, CallMonthly: 0, DurationMonthly: 0}
	//
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"call_monthly": 1, "duration_monthly": cdr.billsec,
			"answere_wait_time": cdr.waitAnswer},
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
			log.Error("[mongo] Monthly insert failed with error : [%v].", err)
			return err
		}
		info, err = collection.Find(selector).Apply(change, &doc)
		if info != nil {
			log.Debugf("[mongo] New record inserted : %s.", doc.Id)
		} else {
			log.Debugf("New record inserted : %s.", doc.Id)
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

func processDailyAnalytics(session *mgo.Session, cdr Cdr) (err error) {
	//
	var collectionName = ""
	if cdr.inoutstatus == DIRECTION_CALL_OUT {
		collectionName = "dailyanalytics_outgoing"
	} else if cdr.inoutstatus == DIRECTION_CALL_IN {
		collectionName = "dailyanalytics_incomming"
	} else {
		return errors.New("[mongo] Can't detect the call context")
	}
	//var t = time.Unix(cdr.calldate, 0)
	var id = fmt.Sprintf("%04d%02d%02d-%s-%d", cdr.calldate.Year(), cdr.calldate.Month(), cdr.calldate.Day(), cdr.src, cdr.causeStatus)
	var metaDate = time.Date(cdr.calldate.Year(), cdr.calldate.Month(), cdr.calldate.Day(), 1, 0, 0, 0, time.UTC)
	log.Debugf("Import daily analytics :  %s for the id %s.", collectionName, id)
	var collection = session.DB(ODIN_MONGO_DATABASENAME).C(collectionName)
	metaDoc := metaData{User: cdr.src, Dt: metaDate, Disposition: cdr.causeStatus}
	doc := dailyCall{Id: id, MetaData: metaDoc, AnswereWaitTime: cdr.waitAnswer, CallDaily: 0, DurationDaily: 0}
	//err = collection.Insert(doc)
	var selector = bson.M{"_id": id, "metadata": metaDoc}
	var hourlyInc = fmt.Sprintf("call_hourly.%d", cdr.calldate.Hour())
	var durationHourlyInc = fmt.Sprintf("duration_hourly.%d", cdr.calldate.Hour())
	//
	var change = mgo.Change{
		Update: bson.M{"$inc": bson.M{"call_daily": 1, "duration_daily": cdr.billsec,
			"answere_wait_time": cdr.waitAnswer, hourlyInc: 1, durationHourlyInc: cdr.billsec},
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

func importCdrToMongo(session *mgo.Session, cdr Cdr) (err error) {
	log.Debugf("Start analyze data for mongo database.")
	createMongoCdr(session, cdr)
	err = processDailyAnalytics(session, cdr)
	if err != nil {
		return err
	}
	err = processMonthlyAnalytics(session, cdr)
	err = nil
	return
}
