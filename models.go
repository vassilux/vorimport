package main

import (
	"labix.org/v2/mgo/bson"
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
	waitAnswer  int
	inoutstatus int
	causeStatus int
}

type Cel struct {
	EventTime int64
}

type RawCall struct {
	Id             bson.ObjectId `bson:"_id"`
	AsteriskId     string        `bson:"asterisk_id"`
	Calldate       time.Time     `bson:"call_date"`
	MetadataDt     time.Time     `bson:"metadata_date"`
	ClidName       string        `bson:"clid_name"`
	ClidNumber     string        `bson:"clid_number"`
	Src            string        `bson:"src"`
	Channel        string        `bson:"channel"`
	Dcontext       string        `bson:"dcontext"`
	DispositionStr string        `bson:"disposition_str"`
	Disposition    int           `bson:"disposition"`
	AnswerWaitTime int           `bson:"answer_wait_time"`
	Billsec        int           `bson:"billsec"`
	Duration       int           `bson:"duration"`
	Uniqueid       string        `bson:"uniqueId"`
	InoutStatus    int           `bson:"inout_status"`
	RecordFile     string        `bson:"record_file"`
	Dst            string        `bson:"dst"`
	Dnid           string        `bson:"dnid"`
	Dstchannel     string        `bson:"dst_channel"`
	CallDetails    []CallDetail  `bson:"call_details"`
}

type CallDetail struct {
	EventType string    `bson:"event_type"`
	EventTime time.Time `bson:"event_time"`
	CidNum    string    `bson:"cid_num"`
	CidDnid   string    `bson:"cid_dnid"`
	Exten     string    `bson:"exten"`
	UniqueId  string    `bson:"uniqueId"`
	LinkedId  string    `bson:"linkedId"`
	Context   string    `bson:"context"`
	Peer      string    `bson:"peer"`
}

type MetaData struct {
	Dst         string    `bson:"dst"`
	Dt          time.Time `bson:"dt"`
	Disposition int       `bson:"disposition"`
}

type DatasByHours struct {
	H0  int `bson:"0"`
	H1  int `bson:"1"`
	H2  int `bson:"2"`
	H3  int `bson:"3"`
	H4  int `bson:"4"`
	H5  int `bson:"5"`
	H6  int `bson:"6"`
	H7  int `bson:"7"`
	H8  int `bson:"8"`
	H9  int `bson:"9"`
	H10 int `bson:"10"`
	H11 int `bson:"11"`
	H12 int `bson:"12"`
	H13 int `bson:"13"`
	H14 int `bson:"14"`
	H15 int `bson:"15"`
	H16 int `bson:"16"`
	H17 int `bson:"17"`
	H18 int `bson:"18"`
	H19 int `bson:"19"`
	H20 int `bson:"20"`
	H21 int `bson:"21"`
	H22 int `bson:"22"`
	H23 int `bson:"23"`
}

type DatasByDayOfMonth struct {
	D1  int `bson:"1"`
	D2  int `bson:"2"`
	D3  int `bson:"3"`
	D4  int `bson:"4"`
	D5  int `bson:"5"`
	D6  int `bson:"6"`
	D7  int `bson:"7"`
	D8  int `bson:"8"`
	D9  int `bson:"9"`
	D10 int `bson:"10"`
	D11 int `bson:"11"`
	D12 int `bson:"12"`
	D13 int `bson:"13"`
	D14 int `bson:"14"`
	D15 int `bson:"15"`
	D16 int `bson:"16"`
	D17 int `bson:"17"`
	D18 int `bson:"18"`
	D19 int `bson:"19"`
	D20 int `bson:"20"`
	D21 int `bson:"21"`
	D22 int `bson:"22"`
	D23 int `bson:"23"`
	D24 int `bson:"24"`
	D25 int `bson:"25"`
	D26 int `bson:"26"`
	D27 int `bson:"27"`
	D28 int `bson:"28"`
	D29 int `bson:"29"`
	D30 int `bson:"30"`
	D31 int `bson:"31"`
}

type DailyCall struct {
	Id              string       `bson:"_id"`
	Meta            MetaData     `bson:"metadata"`
	AnswereWaitTime int          `bson:"answer_wait_time"`
	CallDaily       int          `bson:"call_daily"`
	DurationDaily   int          `bson:"duration_daily"`
	CallsHourly     DatasByHours `bson:"call_hourly"`
	DurationsHourly DatasByHours `bson:"duration_hourly"`
}

type MonthlyCall struct {
	Id              string            `bson:"_id"`
	Meta            MetaData          `bson:"metadata"`
	AnswereWaitTime int               `bson:"answer_wait_time"`
	CallMonthly     int               `bson:"call_monthly"`
	DurationMonthly int               `bson:"duration_monthly"`
	CallsDaily      DatasByDayOfMonth `bson:"calls_daily"`
	DurationsDaily  DatasByDayOfMonth `bson:"durations_daily"`
}
