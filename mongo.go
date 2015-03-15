package main

import (
	log "github.com/cihub/seelog"
	"labix.org/v2/mgo"
	"os"
)

type TestDummy struct {
	FirstName string
}

//
func createMongoCdr(session *mgo.Session, cdr RawCall) (err error) {
	collection := session.DB(config.MongoDbName).C("cdrs")
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

func importCdrToMongo(session *mgo.Session, cdr RawCall) (err error) {
	log.Debugf("Start analyze data for mongo database from asterisk : [%s].", cdr.AsteriskId)
	createMongoCdr(session, cdr)

	if cdr.Dst != "s" {
		//can import for peer(users)
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
	}

	if cdr.InoutStatus == DIRECTION_CALL_IN {
		err = processDidImport(session, cdr)
	}
	err = nil
	return
}
