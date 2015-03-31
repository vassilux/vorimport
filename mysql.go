package main

import (
	"errors"
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
	myQuery := "SELECT UNIX_TIMESTAMP(calldate) as calldate, clid, src, dst, channel, dcontext, disposition,billsec,duration,uniqueid,dstchannel, dnid, recordfile from asteriskcdrdb.cdr WHERE import = 0 and dcontext NOT LIKE 'app-alive-test' LIMIT " + config.DbMySqlFetchRowNumber
	//
	log.Debugf("Executing request [%s]\r\n", myQuery)
	rows, res, err := db.Query(myQuery)
	//
	if err != nil {
		log.Errorf("Executing request [%s] and get error [%s] \r\n", myQuery, err)
		return nil, err
	}
	//
	log.Tracef("getMysqlCdr request executed and get [%d] rows\r\n", len(rows))
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
	myCelQuery := "select UNIX_TIMESTAMP(eventtime) as eventtime from cel where eventtype LIKE 'ANSWER' AND linkedid!=uniqueid AND linkedid='" + uniqueid + "'"
	//
	rows, res, err := db.Query(myCelQuery)
	if err != nil {
		return cel, err
	}

	log.Tracef("getMySqlCel request executed and get [%d] rows\r\n", len(rows))

	if len(rows) > 0 {
		log.Tracef("getMySqlCel get rows for uniqueid [%s]", uniqueid)
		row := rows[0]
		eventtime := res.Map("eventtime")
		cel.EventTime = row.Int64(eventtime)
	} else {
		log.Tracef("getMySqlCel faied to get rows for uniqueid [%s].", uniqueid)
		cel.EventTime = 0
	}

	return cel, nil
}

func getMysqlCdrTestCall(db mysql.Conn) (results []RawCall, err error) {
	log.Tracef("Enter into getMysqlCdr")
	myQuery := "SELECT UNIX_TIMESTAMP(calldate) as calldate, clid, src, dst, channel, dcontext, disposition,billsec,duration,uniqueid,dstchannel, dnid, recordfile from asteriskcdrdb.cdr WHERE import = 0 and dcontext LIKE 'app-alive-test' LIMIT " + config.DbMySqlFetchRowNumber
	//
	log.Debugf("Equery = strings.ToUpper(query)xecuting request [%s]\r\n", myQuery)
	rows, res, err := db.Query(myQuery)
	//
	if err != nil {
		log.Debugf("Executing request [%s] and get error [%s] \r\n", myQuery, err)
		return nil, err
	}
	//
	log.Tracef("getMysqlCdrTestCall Request executed and get [%d] rows\r\n", len(rows))
	//prepare results array
	results = make([]RawCall, len(rows))
	i := 0
	for _, row := range rows {
		//
		var c RawCall //Cdr
		//mappingquery = strings.ToUpper(query) databases fields
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

	var sqlStart = sqlBase + " uniqueid = '" + uniqueid + "' OR linkedid = '" + uniqueid + "' " + sqlOrder

	rows, res, err := db.Query(sqlStart)
	//
	if err != nil {
		log.Errorf("getMySqlCallDetailsExecuting request get error [%s]\n", err)
		return nil, err
	}

	if len(rows) == 0 {
		log.Infof("getMySqlCallDetails 0 records from cel table for uniqueid [%s]\n", uniqueid)
		return nil, nil
	}
	//
	//var searchIdMap = make(map[string]string)
	//for _, row := range rows {
	//	uniqueid := res.Map("uniqueid")
	//	linkedid := res.Map("linkedid")
	//	keyUniqueid := fmt.Sprintf("'%s'", row.Str(uniqueid))
	//	keyLinkedId := fmt.Sprintf("'%s'", row.Str(linkedid))
	//	searchIdMap[keyUniqueid] = keyUniqueid
	//	searchIdMap[keyLinkedId] = keyLinkedId

	//}
	//var keys []string
	//for k := range searchIdMap {
	//	keys = append(keys, k)
	//}
	//var strIds = strings.Join(keys, ",")

	//var sqlNext = sqlBase + "uniqueid IN (" + strIds + ") OR linkedid IN (" + strIds + ")" + sqlOrder
	////

	//log.Tracef("Call details query : [%s]", sqlNext)

	//rowsNext, resNext, errNext := db.Query(sqlNext)
	//if errNext != nil {
	//	log.Errorf(" getMySqlCallDetailsExecuting request [%s] and get error [%s] \r\n", sqlNext, errNext)
	//	return nil, errNext
	//}

	//if len(rowsNext) == 0 {
	//	return nil, nil
	//}

	rowsNext := rows
	resNext := res

	////prepare results array

	results = make([]CallDetail, len(rowsNext))

	i := 0
	for _, rowNext := range rowsNext {
		//
		var c CallDetail
		//mapping databases fields
		eventtype := resNext.Map("eventtype")
		eventtime := resNext.Map("eventtime")
		cid_num := resNext.Map("cid_num")
		cid_dnid := resNext.Map("cid_dnid")
		exten := resNext.Map("exten")
		uniqueid := resNext.Map("uniqueid")
		linkedid := resNext.Map("linkedid")
		context := resNext.Map("context")
		peer := resNext.Map("peer")
		//
		c.EventType = rowNext.Str(eventtype)
		c.EventTime = time.Unix(rowNext.Int64(eventtime)+int64(timeZoneOffset), 0)
		c.CidNum = rowNext.Str(cid_num)
		c.CidDnid = rowNext.Str(cid_dnid)
		c.Exten = rowNext.Str(exten)
		c.UniqueId = rowNext.Str(uniqueid)
		c.LinkedId = rowNext.Str(linkedid)
		c.Peer = rowNext.Str(peer)
		c.Context = rowNext.Str(context)

		results[i] = c
		i++

	}
	log.Tracef("getMySqlCallDetails Return [%d] results .\r\n", len(results))
	return results, nil
}

func udpateMySqlCdrImportStatus(db mysql.Conn, uniqueid string, status int) (err error) {
	var query = fmt.Sprintf("UPDATE cdr SET import = %d WHERE uniqueid = '%s'", status, uniqueid)
	log.Tracef("update cdr status [%s].\n", query)
	_, _, err = db.Query(query)
	//
	return err
}

//Execute customize requests provisted by configuration file
func executeCustomRequests(db mysql.Conn) (err error) {
	for i := range config.CleanupRequests {
		var query = config.CleanupRequests[i]

		if strings.Contains(strings.ToLower(query), "delete") == true {
			return errors.New("Using delete keyword is forbiden. Please check your configuration file(config.json)")
		}
		_, _, err = db.Query(query)

		if err != nil {
			return err
		}

	}
	return nil
}

func deleteMySqlCdrRecord(db mysql.Conn, uniqueid string) (err error) {
	var query = fmt.Sprintf("DELETE FROM cdr WHERE uniqueid = '%s'", uniqueid)
	_, _, err = db.Query(query)
	//
	return err
}

func deleteMySqlCelRecord(db mysql.Conn, uniqueid string) (err error) {

	if config.PurgeCelEvents == false {
		return nil
	}

	var query = fmt.Sprintf("DELETE FROM cel WHERE uniqueid = '%s' OR linkedid = '%s'", uniqueid, uniqueid)
	_, _, err = db.Query(query)

	log.Debugf("getMySqlCallDetails execute : [%s] .\n", query)

	if err != nil {
		log.Errorf("deleteMySqlCelRecord Failed delete record into cel table for uniqueid [%s].\n", uniqueid)
	}
	return err
}
