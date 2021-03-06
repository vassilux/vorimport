package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native"
	"io/ioutil"
	"labix.org/v2/mgo"
	"os"
	"os/signal"
	"redis"
	"sync"
	"syscall"
	"time"
)

const (
	VERSION = "X.X.X"
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
	MongoDbName           string
	EventsMongoHost       string
	AsteriskID            string
	AsteriskAddr          string
	AsteriskPort          int
	AsteriskUser          string
	AsteriskPassword      string
	TestCallActive        bool
	PurgeCelEvents        bool
	TestCallSchedule      int
	DialplanContext       []Context
	Notifications         []string
	Dids                  []string
	ExcludeFromAnalytics  []string
	CleanupRequests       []string
}

var (
	config               *Config
	configLock           = new(sync.RWMutex)
	timeZoneOffset       int64
	isImportProcessing   bool
	configFile           = flag.String("config", "config.json", "Configuration file path")
	importTick           = flag.Int("tick", 10, "Importing tick cycle")
	version              = flag.Bool("version", false, "show version")
	eventWatcher         *EventWatcher
	testCallOriginator   *callOriginator
	stopImportJob        chan bool
	stopGenerateTestCall chan bool
)

const (
	DIRECTION_CALL_OUT    = 1
	DIRECTION_CALL_IN     = 2
	DIRECTION_CALL_IGNORE = 3
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

/**
 *
 */
func loadConfig(fail bool) {
	file, err := ioutil.ReadFile(*configFile)

	if err != nil {
		log.Errorf("Can't open configuration file : %s", err)
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
	if len(config.MongoDbName) == 0 {
		config.MongoDbName = "revor"
	}
	configLock.Unlock()
}

func loadLogger() {
	logger, err := log.LoggerFromConfigAsFile("logger.xml")

	if err != nil {
		log.Error("Can not load the logger configuration file, Please check if the file logger.xml exists on current directory", err)
		os.Exit(1)
	} else {
		log.ReplaceLogger(logger)
		logger.Flush()
	}

}

func GetConfig() *Config {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func init() {
	//

}

func getInOutStatus(cdr RawCall) (status int, err error) {
	return getInOutContextStatus(cdr.Dcontext)
}

func getInOutContextStatus(context string) (status int, err error) {

	for i := range config.DialplanContext {
		if config.DialplanContext[i].Name == context {
			status = config.DialplanContext[i].Direction
			return status, nil

		}
	}
	log.Infof("Can not find the call direction for the context [%s].", context)
	return status, errors.New("Can't find the context direction for the context : " + context)

}

func syncPublish(spec *redis.ConnectionSpec, channel string, messageType string) {

	client, err := redis.NewSynchClientWithSpec(spec)
	if err != nil {
		log.Errorf("Failed to create the redis client : %s", err)
		os.Exit(1)
	}

	msg := []byte(fmt.Sprintf("{id : %s }", messageType))
	rcvCnt, err := client.Publish(channel, msg)
	if err != nil {
		log.Errorf("Error to publish the messge to the redis : %s", err)
	} else {
		log.Debugf("Message published to %d subscribers", rcvCnt)
	}

	client.Quit()
}

func sendEventNotification(flag int, datas string) {
	ev := &Event{
		Mask:  new(BitSet),
		Datas: datas,
	}
	ev.Mask.Set(flag)
	eventWatcher.event <- ev
}

func sendMySqlEventNotification(flag int) {
	datas := fmt.Sprintf("MySql server : %s change state", config.DbMySqlHost)
	sendEventNotification(flag, datas)
}

func sendMongoEventNotification(flag int) {
	datas := fmt.Sprintf("Mongo server : %s change state", config.DbMySqlHost)
	sendEventNotification(flag, datas)
}

func importJob() {
	//
	db := mysql.New("tcp", "", config.DbMySqlHost, config.DbMySqlUser, config.DbMySqlPassword, config.DbMySqlName)
	log.Debugf("Connecting to the database %s %s %s %s.", config.DbMySqlHost, config.DbMySqlUser, config.DbMySqlPassword, config.DbMySqlName)
	//
	err := db.Connect()
	if err != nil {
		sendMySqlEventNotification(MYSQKO)
		log.Criticalf("Can't connect to the mysql database error : %s.", err)
		return
	}
	sendMySqlEventNotification(MYSQOK)
	log.Debug("Connected to the mysql database with success.")
	//
	session, err := mgo.Dial(config.MongoHost)
	if err != nil {
		log.Debugf("Can't connect to the mongo database error : %s.", err)
		sendMongoEventNotification(MONGKO)
		return
	}
	session.SetMode(mgo.Monotonic, true)
	defer session.Close()
	sendMongoEventNotification(MONGOK)
	log.Debug("Connected to the mongo database with success.")
	//
	err = executeCustomRequests(db)
	if err != nil {
		log.Errorf("Error to execute customise request : %s", err)
		log.Flush()
		os.Exit(1)
	}
	//
	cdrs, err := getMysqlCdr(db)
	//
	if err != nil {
		log.Criticalf("Can not get records from mysql cause error [%s].", err)
		log.Flush()
		os.Exit(1)
	}
	log.Tracef("Start records parcing.")
	//
	var incommingCount = 0
	var outgoingCount = 0

	for _, cdr := range cdrs {
		cdr.AsteriskId = config.AsteriskID
		var datetime = cdr.Calldate.Format(time.RFC3339)
		log.Tracef("Get raw cdr for the date [%s], the clid [%s] and the context [%s] from asterisk [%s]", datetime, cdr.ClidNumber, cdr.Dcontext, cdr.AsteriskId)
		//var cel Cel
		//cel, err = getMySqlCel(db, cdr.Uniqueid)
		var inoutstatus, err = getInOutStatus(cdr)
		if err != nil {
			log.Criticalf("Get error[%s]. Please check your configuration file.", err)
			log.Flush()
			os.Exit(1)
		}
		if inoutstatus == 1 {
			outgoingCount++
		} else if inoutstatus == 2 {
			incommingCount++
		}
		cdr.InoutStatus = inoutstatus
		var dispostionCode = DIC_DISPOSITION[cdr.DispositionStr]

		if dispostionCode > 0 {
			cdr.Disposition = DISPOSITION_TRANSLATION[dispostionCode]
		} else {
			cdr.Disposition = 0
		}

		//if cel.EventTime > 0 {
		//	//extract the timezone offset
		//	cdr.AnswerWaitTime = int(cel.EventTime - cdr.Calldate.Unix() - timeZoneOffset)
		//}
		//
		callDetails, err := getMySqlCallDetails(db, cdr.Uniqueid)
		if err != nil {
			log.Criticalf("Try to get the call details but get the error[%s].", err)
			log.Flush()
			os.Exit(1)
		}
		//
		log.Tracef("Get [%d] details records for the call with uniqueid [%s].", len(callDetails), cdr.Uniqueid)
		if callDetails != nil {
			cdr.CallDetails = callDetails
		}
		//

		if cdr.InoutStatus == DIRECTION_CALL_IN {
			var did = ""
			var peer = ""
			var exten = ""

			if len(cdr.CallDetails) == 0 && cdr.Dnid != "" {
				//a workaround if I can get cel records
				did = cdr.Dnid
				log.Tracef("Force did from cdr dnid [%s].", did)
			}

			for i := range cdr.CallDetails {
				var callDetail = cdr.CallDetails[i]
				if i == 0 && callDetail.EventType == "CHAN_START" {
					exten = callDetail.Exten
					did = callDetail.Exten
				}

				if callDetail.EventType == "ANSWER" && exten == callDetail.CidDnid && did == "" {
					//try to find did
					did = callDetail.CidDnid

				} else if callDetail.EventType == "BRIDGE_START" {
					//bridge start gives the time before answer
					cdr.AnswerWaitTime = int(callDetail.EventTime.Unix() - cdr.Calldate.Unix())
					peer = getPeerFromChannel(callDetail.Peer)

				} else if callDetail.EventType == "BRIDGE_END" && peer == "" {
					//idea to find the last BRIDGE_END event and get the extention from it
					peer = getPeerFromChannel(callDetail.Peer)
					break
				}
			}

			if peer == "" {
				peer = getPeerFromChannel(cdr.Dstchannel)
			}

			if peer == "" {
				//can be case that the call is not answered
				peer = cdr.Dst
			}

			cdr.Peer = peer

			//there is a possibility to have for incomming call a peer for the did number
			//this is dependes of the customer configuration
			//try to find a did value from database defined by the customer
			err := isDid(session, did)

			if err == nil {
				cdr.Did = did
			} else {
				cdr.Did = ""
			}

			log.Tracef("Did [%s] and peer [%s] for the call with uniqueid [%s].\n", did, peer, cdr.Uniqueid)

			if cdr.Dst == "s" && peer != "" {
				//
				cdr.Dst = cdr.Peer
			}

		} //end of the incomming call process
		//
		//
		err = importCdrToMongo(session, cdr)
		var importedStatus = 1
		if err != nil {
			log.Errorf("Can't import cdr to mongo [%v].", err)
			log.Flush()
			os.Exit(1)
		}
		//
		log.Debugf("Import executed for unique id [%s] with code : [%d], try process the mysql updating.\n",
			cdr.Uniqueid, importedStatus)

		err = udpateMySqlCdrImportStatus(db, cdr.Uniqueid, 1)

		if err != nil {
			log.Errorf("Can't update the import status for the call with unique id [%s].", cdr.Uniqueid)
			log.Flush()
			os.Exit(1)
		}

		err = deleteMySqlCelRecord(db, cdr.Uniqueid)
		if err != nil {
			log.Errorf("Can't delete cel records for uniqueid[%s].", cdr.Uniqueid)
		}

	}
	//
	log.Trace("End of cdr parsing.\n")
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
	//
}

func cleanup() {
	stopImportJob <- true
	if config.TestCallActive == true {
		//wait testing call process
		stopGenerateTestCall <- true
	}

	//
	sendEventNotification(APPSTO, "Application stopped")
	//wait for the eventWatcher
	select {
	case <-eventWatcher.done:
		log.Info("Event watcher stopped.")
	}
	log.Info("Execute the application cleanup")
	log.Flush()
}

func generateTestCall() {
	//

	testCallOriginator.testCall <- true
	//
	select {
	case res := <-testCallOriginator.resultTestCall:
		if res != nil {
			data := fmt.Sprintf("Connection failed to the asterisk server %s.", res)
			sendEventNotification(TCALKO, data)
			return
		} else {
			sendEventNotification(TCALOK, "Test call ok")
		}
	}
	//little stupid wait
	time.Sleep(3 * time.Second)
	db := mysql.New("tcp", "", config.DbMySqlHost, config.DbMySqlUser, config.DbMySqlPassword, config.DbMySqlName)
	log.Debugf("Connecting to the database %s %s %s %s.", config.DbMySqlHost, config.DbMySqlUser, config.DbMySqlPassword, config.DbMySqlName)
	//
	err := db.Connect()
	if err != nil {
		data := fmt.Sprintf("Failed generate test call : %v.", err)
		sendEventNotification(CCALKO, data)
		return
	}
	//
	cdrs, err := getMysqlCdrTestCall(db)
	if len(cdrs) == 0 {
		data := fmt.Sprintf("Failed find a test call for the asterisk server %s.", config.AsteriskID)
		sendEventNotification(CCALKO, data)
		log.Errorf(data)
	} else {
		for _, cdr := range cdrs {
			err = deleteMySqlCdrRecord(db, cdr.Uniqueid)
			if err != nil {
				log.Errorf("Can't delete the test call record with unique id [%s] cause get an error %v.", cdr.Uniqueid, err)
				cleanup()
				os.Exit(1)
			}
		}
		data := fmt.Sprintf("Check call ok for the asterisk server %s.", config.AsteriskID)
		sendEventNotification(CCALOK, data)
		log.Infof("Asterisk the test call processed with success.")
	}
}

func checkConfigAndDie() {
	if len(config.AsteriskAddr) == 0 {
		log.Critical("Asterisk address is null or empty. Please check the configuraiton file.\n")
		log.Flush()
		os.Exit(1)
	}

	if len(config.AsteriskUser) == 0 || len(config.AsteriskPassword) == 0 {
		log.Critical("Asterisk credentials missing.  Please check the configuration file.\n")
		log.Flush()
		os.Exit(1)
	}

	if config.TestCallActive && config.TestCallSchedule < 30 {
		log.Criticalf("Asterisk tesing interval is too short : %d. Minimal value is 30(seconds).Please check the configuration file.\n", config.TestCallSchedule)
		log.Flush()
		os.Exit(1)
	}
}

func main() {
	flag.Parse()
	//
	if *version {
		fmt.Printf("Version : %s\n", VERSION)
		fmt.Println("Get fun! Live well !")
		return
	}
	//
	loadLogger()
	loadConfig(true)
	//
	config = GetConfig()
	//to be carrefule and check if the configuraiton is not
	checkConfigAndDie()

	//
	eventWatcher = NewEventWatcher(config)
	go eventWatcher.run()
	//
	if config.TestCallSchedule > 0 {
		log.Infof("Create test call originator for the asterisk %s:%d", config.AsteriskAddr, config.AsteriskPort)
		testCallOriginator = NewCallOriginator(config.AsteriskAddr, config.AsteriskPort, config.AsteriskUser, config.AsteriskPassword)
	}

	//something wrong I cannot trup SIGUSR1 :-)
	/*s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGUSR1)*/

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		cleanup()
		os.Exit(1)
		/*<-s
		loadConfig(false)
		log.Info("Configuration reloading")*/
	}()
	//
	log.Infof("Starting for %s", config.AsteriskID)
	//dummy flag for indicate that the import is processing
	isImportProcessing = false
	//
	now := time.Now()
	_, timeZoneOffset := now.Zone()
	log.Infof("Startring and using the timezone offset used : %d.", timeZoneOffset)
	//go for schedule job
	duration := time.Duration(*importTick) * time.Second
	//ticker := time.NewTicker(duration)
	stopImportJob = schedule(importJob, duration)
	sendEventNotification(APPSTA, "Application vorimport started.")

	durationTestCall := time.Duration(config.TestCallSchedule) * time.Second
	//ticker := time.NewTicker(duration)
	if config.TestCallActive {
		stopGenerateTestCall = schedule(generateTestCall, durationTestCall)
	}

	for {
		log.Debug("Working...")
		time.Sleep(1000 * time.Second) //

	}

}
