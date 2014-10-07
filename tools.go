package main

import (
	"strings"
	"time"
)

func getPeerFromChannel(channel string) (peer string) {
	w := strings.FieldsFunc(channel, func(r rune) bool {
		switch r {
		case '/', '-', ' ':
			return true
		}
		return false
	})

	if len(w) >= 3 {
		return w[len(w)-2]
	} else {
		return channel
	}
}

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}
