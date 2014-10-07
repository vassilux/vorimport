package main

import (
	gami "code.google.com/p/gami"
	"fmt"
	"log"
	"net"
)

type callOriginator struct {
	Addr     string
	Port     int
	testCall chan bool
	Username string
	Password string
}

func NewCallOriginator(addr string, port int, user string, pswd string) *callOriginator {
	//
	originator := &callOriginator{
		Addr:     addr,
		Port:     port,
		Username: user,
		Password: pswd,
		testCall: make(chan bool, 1),
	}

	go originator.run()
	return originator
}

func (originator *callOriginator) processTestCall() {
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", originator.Addr, originator.Port))

	if err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	g := gami.NewAsterisk(&c, nil)

	err = g.Login(originator.Username, originator.Password)

	if err != nil {
		log.Fatal(err)
	}

	ch := "Local/testcall@app-alive-test"

	o := gami.NewOriginateApp(ch, "System", "touch /tmp/vorimport")

	cb := func(m gami.Message) {
		fmt.Println(m)
	}
	err = g.Originate(o, nil, &cb)

	if err != nil {
		log.Fatal(err)
	}

}

func (originator *callOriginator) run() {
	for {
		select {
		case c := <-originator.testCall:
			if c == true {
				originator.processTestCall()
			}

		}
	}
}
