package main

import (
	"strings"
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
