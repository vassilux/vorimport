package main

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	"labix.org/v2/mgo/bson"
	"strings"
	"time"
)

/**
 * Used for split clid into two string caller name and caller number
 */
func bracket(r rune) bool {
	return r == '<' || r == '>'
}

/**
 *
 */
func getMysqlCdr(db mysql.Conn) (results []RawCall, err error) {
	log.Tracef("Enter into getMysqlCdr")
	myQuery := "SELECT UNIX_TIMESTAMP(calldate) as calldate, clid, src, dst, channel, dcontext, disposition,billsec,duration,uniqueid,dstchannel, dnid, recordfile from asteriskcdrdb.cdr WHERE import = 0 and dcontext NOT LIKE 'app-alive-test' LIMIT 0, " + config.DbMySqlFetchRowNumber
	//
	log.Tracef("Executing request [%s]\r\n", myQuery)
	rows, res, err := db.Query(myQuery)
	//
	if err != nil {
		log.Debugf("Executing request [%s] and get error [%s] \r\n", myQuery, err)
		return nil, err
	}
	//
	log.Tracef("Request executed and get [%d] rows\r\n", len(rows))
	//prepare results array
	results = make([]RawCall, len(rows))
	i := 0
	for _, row := range rows {
		//
		var c RawCall //Cdr
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
		dnid := res.Map("dnid")
		recordfile := res.Map("recordfile")
		dstchannel := res.Map("dstchannel")
		//
		raw_clid := strings.FieldsFunc(row.Str(clid), bracket)
		caller_name := ""
		caller_number := ""

		if len(raw_clid) == 2 {
			caller_name = raw_clid[0]
			caller_number = raw_clid[1]
		} else if len(raw_clid) == 1 {
			caller_number = raw_clid[0]
		}

		/*if len(raw_clid) == 2 {
			caller_name := raw_clid[0]
			caller_number := raw_clid[1]

		}else len(raw_clid) == 1 {
			caller_number := raw_clid[0]
		}*/
		//
		c = RawCall{Id: bson.NewObjectId(),
			Calldate:       time.Unix(row.Int64(calldate)+int64(timeZoneOffset), 0),
			MetadataDt:     time.Unix(time.Now().Unix()+int64(timeZoneOffset), 0),
			ClidName:       caller_name,
			ClidNumber:     caller_number,
			Src:            row.Str(src),
			Channel:        row.Str(channel),
			Dcontext:       row.Str(dcontext),
			DispositionStr: row.Str(disposition),
			Disposition:    0,
			AnswerWaitTime: 0,
			Billsec:        row.Int(billsec),
			Duration:       row.Int(duration),
			Uniqueid:       row.Str(uniqueid),
			InoutStatus:    0,
			RecordFile:     row.Str(recordfile),
			Dst:            row.Str(dst),
			Dnid:           row.Str(dnid),
			Dstchannel:     row.Str(dstchannel)}
		//
		results[i] = c
		i++

	}
	return results, nil
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
		log.Tracef("Get rows for uniqueid [%s]", uniqueid)
		row := rows[0]
		eventtime := res.Map("eventtime")
		cel.EventTime = row.Int64(eventtime)
	} else {
		log.Tracef("Can't get rows for uniqueid [%s].", uniqueid)
		cel.EventTime = 0
	}

	return cel, nil
}

func getMysqlCdrTestCall(db mysql.Conn) (results []RawCall, err error) {
	log.Tracef("Enter into getMysqlCdr")
	myQuery := "SELECT UNIX_TIMESTAMP(calldate) as calldate, clid, src, dst, channel, dcontext, disposition,billsec,duration,uniqueid,dstchannel, dnid, recordfile from asteriskcdrdb.cdr WHERE import = 0 and dcontext LIKE 'app-alive-test' LIMIT 0, " + config.DbMySqlFetchRowNumber
	//
	log.Tracef("Executing request [%s]\r\n", myQuery)
	rows, res, err := db.Query(myQuery)
	//
	if err != nil {
		log.Debugf("Executing request [%s] and get error [%s] \r\n", myQuery, err)
		return nil, err
	}
	//
	log.Tracef("Request executed and get [%d] rows\r\n", len(rows))
	//prepare results array
	results = make([]RawCall, len(rows))
	i := 0
	for _, row := range rows {
		//
		var c RawCall //Cdr
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
		dnid := res.Map("dnid")
		recordfile := res.Map("recordfile")
		dstchannel := res.Map("dstchannel")
		//
		raw_clid := strings.FieldsFunc(row.Str(clid), bracket)
		caller_name := ""
		caller_number := ""

		if len(raw_clid) == 2 {
			caller_name = raw_clid[0]
			caller_number = raw_clid[1]
		} else if len(raw_clid) == 1 {
			caller_number = raw_clid[0]
		}

		/*if len(raw_clid) == 2 {
			caller_name := raw_clid[0]
			caller_number := raw_clid[1]

		}else len(raw_clid) == 1 {
			caller_number := raw_clid[0]
		}*/
		//
		c = RawCall{Id: bson.NewObjectId(),
			Calldate:       time.Unix(row.Int64(calldate)+int64(timeZoneOffset), 0),
			MetadataDt:     time.Unix(time.Now().Unix()+int64(timeZoneOffset), 0),
			ClidName:       caller_name,
			ClidNumber:     caller_number,
			Src:            row.Str(src),
			Channel:        row.Str(channel),
			Dcontext:       row.Str(dcontext),
			DispositionStr: row.Str(disposition),
			Disposition:    0,
			AnswerWaitTime: 0,
			Billsec:        row.Int(billsec),
			Duration:       row.Int(duration),
			Uniqueid:       row.Str(uniqueid),
			InoutStatus:    0,
			RecordFile:     row.Str(recordfile),
			Dst:            row.Str(dst),
			Dnid:           row.Str(dnid),
			Dstchannel:     row.Str(dstchannel)}
		//
		results[i] = c
		i++

	}
	return results, nil
}

func getMySqlCallDetails(db mysql.Conn, uniqueid string) (results []CallDetail, err error) {
	var sqlBase = "SELECT eventtype, UNIX_TIMESTAMP(eventtime) as eventtime, cid_num,  cid_dnid, exten, context, peer, uniqueid, linkedid  FROM  cel WHERE "
	var sqlOrder = " order by eventtime, id"
	//
	var sqlStart = sqlBase + "uniqueid = " + uniqueid + " OR linkedid = " + uniqueid + sqlOrder
	//
	/*myQuery := "SELECT eventtype, UNIX_TIMESTAMP(eventtime) as eventtime, cid_num,  cid_dnid, exten, context, peer, uniqueid, linkedid  FROM  cel WHERE uniqueid =" +
	uniqueid + " OR linkedid = " + uniqueid + " order by eventtime, id"*/
	log.Tracef(" getMySqlCallDetails Executing request [%s]\r\n", sqlStart)
	rows, res, err := db.Query(sqlStart)
	//
	if err != nil {
		log.Debugf(" getMySqlCallDetailsExecuting request [%s] and get error [%s] \r\n", sqlStart, err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	//
	var searchIdMap = make(map[string]string)
	for _, row := range rows {
		uniqueid := res.Map("uniqueid")
		linkedid := res.Map("linkedid")
		searchIdMap[row.Str(uniqueid)] = row.Str(uniqueid)
		searchIdMap[row.Str(linkedid)] = row.Str(linkedid)

	}
	var keys []string
	for k := range searchIdMap {
		keys = append(keys, k)
	}
	var strIds = strings.Join(keys, ",")

	var sqlNext = sqlBase + "uniqueid IN (" + strIds + ") OR linkedid IN (" + strIds + ")" + sqlOrder
	//
	rows, res, err = db.Query(sqlNext)
	if err != nil {
		log.Debugf(" getMySqlCallDetailsExecuting request [%s] and get error [%s] \r\n", sqlNext, err)
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	//prepare results array
	results = make([]CallDetail, len(rows))
	log.Debugf("getMySqlCallDetails create results  for [%d] rows\r\n", len(rows))
	i := 0
	for _, row := range rows {
		//
		var c CallDetail
		//mapping databases fields
		eventtype := res.Map("eventtype")
		eventtime := res.Map("eventtime")
		cid_num := res.Map("cid_num")
		cid_dnid := res.Map("cid_dnid")
		exten := res.Map("exten")
		uniqueid := res.Map("uniqueid")
		linkedid := res.Map("linkedid")
		context := res.Map("context")
		peer := res.Map("peer")
		//
		c.EventType = row.Str(eventtype)
		c.EventTime = time.Unix(row.Int64(eventtime)+int64(timeZoneOffset), 0)
		c.CidNum = row.Str(cid_num)
		c.CidDnid = row.Str(cid_dnid)
		c.Exten = row.Str(exten)
		c.UniqueId = row.Str(uniqueid)
		c.LinkedId = row.Str(linkedid)
		c.Peer = row.Str(peer)
		c.Context = row.Str(context)

		results[i] = c
		i++

	}
	log.Debugf("getMySqlCallDetails Return [%d] results .\r\n", len(results))
	return results, nil
}

func udpateMySqlCdrImportStatus(db mysql.Conn, uniqueid string, status int) (err error) {
	var query = fmt.Sprintf("UPDATE cdr SET import = %d WHERE uniqueid = '%s'", status, uniqueid)
	_, _, err = db.Query(query)
	//
	return err
}

func deleteMySqlCdrRecord(db mysql.Conn, uniqueid string) (err error) {
	var query = fmt.Sprintf("DELETE FROM cdr WHERE uniqueid = '%s'", uniqueid)
	_, _, err = db.Query(query)
	//
	return err
}
