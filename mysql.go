package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"time"
)

type Cdr struct {
	calldate    time.Time
	clid        string
	src         string
	dst         string
	channel     string
	dcontext    string
	disposition string
	billsec     int
	duration    int
	uniqueid    string
	dstchannel  string
	recordfile  string
	waitAnswer  int64
	inoutstatus int
	causeStatus int
}

type Cel struct {
	eventtime int64
}

/**
 *
 */
func getMysqlCdr(db mysql.Conn) (results []Cdr, err error) {
	myQuery := "SELECT UNIX_TIMESTAMP(calldate) as calldate, clid, src, dst, channel, dcontext, disposition,billsec,duration,uniqueid,dstchannel,recordfile from cdr WHERE import = 0 LIMIT 0, " + config.DbMySqlFetchRowNumber
	//
	rows, res, err := db.Query(myQuery)
	if err != nil {
		return nil, err
	}
	//prepare results array
	results = make([]Cdr, len(rows))
	i := 0
	for _, row := range rows {
		//
		var c Cdr
		//mapping databases fields
		calldate := res.Map("calldate")
		clid := res.Map("clid")
		src := res.Map("src")
		dst := res.Map("dst")
		channel := res.Map("channel")
		dcontext := res.Map("dcontext")
		disposition := res.Map("disposition")
		billsec := res.Map("billsec")
		duration := res.Map("duration")
		uniqueid := res.Map("uniqueid")
		dstchannel := res.Map("dstchannel")
		recordfile := res.Map("recordfile")
		//
		c.calldate = time.Unix(row.Int64(calldate)+int64(timeZoneOffset), 0)
		c.clid = row.Str(clid)
		c.src = row.Str(src)
		c.dst = row.Str(dst)
		c.channel = row.Str(channel)
		c.dcontext = row.Str(dcontext)
		c.disposition = row.Str(disposition)
		c.billsec = row.Int(billsec)
		c.duration = row.Int(duration)
		c.uniqueid = row.Str(uniqueid)
		c.dstchannel = row.Str(dstchannel)
		c.recordfile = row.Str(recordfile)
		//
		results[i] = c
		i++

	}
	return
}

/**
 * Process selection CEL events for given unique id
 */
func getMySqlCel(db mysql.Conn, uniqueid string) (cel Cel, err error) {
	myCelQuery := "select UNIX_TIMESTAMP(eventtime) as eventtime from cel where eventtype LIKE 'ANSWER' AND linkedid!=uniqueid AND linkedid=" + uniqueid
	//
	rows, res, err := db.Query(myCelQuery)
	if err != nil {
		return cel, err
	}
	if len(rows) > 0 {
		log.Debugf("Get rows for uniqueid %d", uniqueid)
		row := rows[0]
		eventtime := res.Map("eventtime")
		cel.eventtime = row.Int64(eventtime)
	} else {
		log.Infof("Can't get rows for uniqueid %d.", uniqueid)
		cel.eventtime = 0
	}

	return cel, err
}

func udpateMySqlCdrImportStatus(db mysql.Conn, uniqueid string, status int) (err error) {
	var query = fmt.Sprintf("UPDATE cdr SET import = %d WHERE uniqueid = '%s'", status, uniqueid)
	_, _, err = db.Query(query)
	//
	return err
}
