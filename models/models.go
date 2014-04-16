package models

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
	Calldate       time.Time     `bson:"calldate"`
	MetadataDt     time.Time     `bson:"metadataDt"`
	ClidName       string        `bson:"clidName"`
	ClidNumber     string        `bson:"clidNumber"`
	Src            string        `bson:"src"`
	Channel        string        `bson:"channel"`
	Dcontext       string        `bson:"dcontext"`
	DispositionStr string        `bson:"dispositionStr"`
	Disposition    int           `bson:"disposition"`
	AnswerWaitTime int           `bson:"answerwaittime"`
	Billsec        int           `bson:"billsec"`
	Duration       int           `bson:"duration"`
	Uniqueid       string        `bson:"uniqueid"`
	InoutStatus    int           `bson:"inoutstatus"`
	RecordFile     string        `bson:"recordfile"`
	Dst            string        `bson:"dst"`
	Dnid           string        `bson:"dnid"`
	Dstchannel     string        `bson:"dstchannel"`
	CallDetails    []CallDetail  `bson:"callDetails"`
}

type CallDetail struct {
	EventType string    `bson:"eventType"`
	EventTime time.Time `bson:"eventTime"`
	CidNum    string    `bson:"cidNum"`
	CidDnid   string    `bson:"cidDnid"`
	Exten     string    `bson:"exten"`
	UniqueId  string    `bson:"uniqueId"`
	LinkedId  string    `bson:"linkedId"`
	Peer      string    `bson:"peer"`
}

type MetaData struct {
	User        string    `bson:"user"`
	Dt          time.Time `bson:"dt"`
	Disposition int       `bson:"disposition"`
}

type DailyCall struct {
	Id              string   `bson:"_id"`
	Meta            MetaData `bson:"metadata"`
	AnswereWaitTime int      `bson:"answere_wait_time"`
	CallDaily       int      `bson:"call_daily"`
	DurationDaily   int      `bson:"duration_daily"`
}

type MonthlyCall struct {
	Id              string   `bson:"_id"`
	Meta            MetaData `bson:"metadata"`
	AnswereWaitTime int      `bson:"answere_wait_time"`
	CallMonthly     int      `bson:"call_monthly"`
	DurationMonthly int      `bson:"duration_monthly"`
}
