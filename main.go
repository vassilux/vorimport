package main

import (
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine
	// _ "github.com/ziutek/mymysql/thrsafe" // Thread safe engine
	"labix.org/v2/mgo"
	//"labix.org/v2/mgo/bson"
	//"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"io/ioutil"
	stdlog "log"
	"os"
	"os/signal"
	"redis"
	"sync"
	"syscall"
	"time"
)

type Context struct {
	Name      string
	Direction int
}

type Config struct {
	DbMySqlHost           string
	DbMySqlUser           string
	DbMySqlPassword       string
	DbMySqlName           string
	DbMySqlFetchRowNumber string
	MongoHost             string
	DialplanContext       []Context
}

var (
	config         *Config
	configLock     = new(sync.RWMutex)
	timeZoneOffset int64
)

const (
	DIRECTION_CALL_OUT = 1
	DIRECTION_CALL_IN  = 2
)

var (
	DISPOSITION_TRANSLATION map[int]int = map[int]int{
		0:  0,
		1:  16, // ANSWER
		2:  17, // BUSY
		3:  19, // NOANSWER
		4:  21, // CANCEL
		5:  34, // CONGESTION
		6:  47, // CHANUNAVAIL
		7:  0,  // DONTCALL
		8:  0,  // TORTURE
		9:  0,  // INVALIDARGS
		10: 41, // FAILED
	}
)

var (
	DIC_DISPOSITION map[string]int = map[string]int{
		"ANSWER":      1,
		"ANSWERED":    1,
		"BUSY":        2,
		"NOANSWER":    3,
		"NO ANSWER":   3,
		"CANCEL":      4,
		"CONGESTION":  5,
		"CHANUNAVAIL": 6,
		"DONTCALL":    7,
		"TORTURE":     8,
		"INVALIDARGS": 9,
		"FAIL":        10,
		"FAILED":      10,
	}
)

var log = logging.MustGetLogger("vorimport")

/**
 *
 */
func loadConfig(fail bool) {
	file, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Error("Can't open configuration file : %s", err)
		if fail {
			os.Exit(1)
		}
	}

	temp := new(Config)
	if err = json.Unmarshal(file, temp); err != nil {
		log.Error("Can't load json configuration : %s", err)
		if fail {
			os.Exit(1)
		}
	}
	configLock.Lock()
	config = temp
	configLock.Unlock()
}

func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func init() {
	//called on the start by go
	loadConfig(true)
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR2)
	go func() {
		for {
			<-s
			loadConfig(false)
			log.Error("Configuration relaoding")
		}
	}()
}

func getInOutStatus(cdr Cdr) (status int, err error) {
	config = GetConfig()
	for i := range config.DialplanContext {
		if config.DialplanContext[i].Name == cdr.dcontext {
			status = config.DialplanContext[i].Direction
			return status, nil

		}
	}
	return status, errors.New("[main] Can't find the context direction")

}

func syncPublish(spec *redis.ConnectionSpec, channel string, messageType string) {

	client, err := redis.NewSynchClientWithSpec(spec)
	if err != nil {
		log.Error("Failed to create the redis client : %s", err)
		os.Exit(1)
	}

	msg := []byte(fmt.Sprintf("{id : %s }", messageType))
	rcvCnt, err := client.Publish(channel, msg)
	if err != nil {
		log.Error("Error to publish the messge to the redis : %s", err)
	} else {
		log.Debug("Message published to %d subscribers", rcvCnt)
	}

	client.Quit()
}

func importJob() {
	//
	db := mysql.New("tcp", "", config.DbMySqlHost, config.DbMySqlUser, config.DbMySqlPassword, config.DbMySqlName)
	log.Debug("Connecting to the database %s %s %s %s.", config.DbMySqlHost, config.DbMySqlUser, config.DbMySqlPassword, config.DbMySqlName)
	//
	err := db.Connect()
	if err != nil {
		log.Debug("Can't connect to the mysql database error : %s.", err)
		os.Exit(1)
	}
	log.Debug("Connected to the mysql database with success.")
	//
	session, err := mgo.Dial(config.MongoHost)
	if err != nil {
		log.Debug("Can't connect to the mongo database error : %s.", err)
		os.Exit(1)
	}
	session.SetMode(mgo.Monotonic, true)
	defer session.Close()
	log.Debug("Connected to the mongo database with success.")
	//
	cdrs, err := getMysqlCdr(db)
	//
	if err != nil {
		panic(err)
	}
	//
	var incommingCount = 0
	var outgoingCount = 0
	for _, cdr := range cdrs {
		var datetime = cdr.calldate.Format(time.RFC3339)
		log.Debug("Get raw cdr for the date %s the clid % and the context %s", datetime, cdr.clid, cdr.dcontext)
		var cel Cel
		cel, err = getMySqlCel(db, cdr.uniqueid)
		var inoutstatus, err = getInOutStatus(cdr)
		if err != nil {
			log.Debug("Can't detect direction of the context %s", cdr.dcontext)
			os.Exit(1)
		}
		if inoutstatus == 1 {
			outgoingCount++
		} else if inoutstatus == 2 {
			incommingCount++
		}
		cdr.inoutstatus = inoutstatus
		var dispostionCode = DIC_DISPOSITION[cdr.disposition]

		if dispostionCode > 0 {
			cdr.causeStatus = DISPOSITION_TRANSLATION[dispostionCode]
		} else {
			cdr.causeStatus = 0
		}

		if cel.eventtime > 0 {
			//extract the timezone offset
			cdr.waitAnswer = cel.eventtime - (cdr.calldate.Unix() - timeZoneOffset)
		}

		err = importCdrToMongo(session, cdr)
		var importedStatus = 1
		if err != nil {
			importedStatus = -1
		}
		//
		log.Info("Import executed for unique id [%s] with code : [%d], try process the mysql process updating.\n", cdr.uniqueid, importedStatus)
		err = udpateMySqlCdrImportStatus(db, cdr.uniqueid, importedStatus)
		if err != nil {
			log.Error("Can't update the import status for the call with unique id [%s].", cdr.uniqueid)
			os.Exit(1)
		}
	}
	//
	spec := redis.DefaultSpec()
	channel := "channel_cdr"
	//
	if incommingCount > 0 {
		syncPublish(spec, channel, "cdrincomming")
	}

	if outgoingCount > 0 {
		syncPublish(spec, channel, "cdroutgoing")
	}

}

func cleanup() {
	log.Info("Execute the application cleanup")
}

func main() {
	var format = logging.MustStringFormatter("%{level} %{message}")
	logging.SetFormatter(format)
	logging.SetLevel(logging.INFO, "package.example")

	logBackend := logging.NewLogBackend(os.Stderr, "", stdlog.LstdFlags|stdlog.Lshortfile)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	//
	config = GetConfig()
	//
	now := time.Now()
	_, timeZoneOffset := now.Zone()
	log.Info("Use the timezone offset used : %d.", timeZoneOffset)
	//
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		<-c
		cleanup()
		os.Exit(1)
	}()
	//
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				importJob()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	for {
		log.Debug("Sleeping...")
		time.Sleep(10 * time.Second) // placeholder for the future can be used for the application state check

	}

}
