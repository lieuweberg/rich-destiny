package main

import (
	"log"
	"runtime"
	"strings"
	"time"
)

var spamlessErrors = make(map[errorOrigin][]spamlessError)

type spamlessError struct {
	msg      string
	amount   int
	lastSeen int64
}

func logErrorIfNoErrorSpam(origin errorOrigin, msg string) {
	if o, ok := spamlessErrors[origin]; ok {
		for i := 0; i < len(o); i++ {
			e := o[i]

			// Clear errors that weren't seen for a while. New errors will check if every error in the origin needs to be cleared.
			if time.Now().Unix()-e.lastSeen > 5*60 {
				o[i] = o[len(o)-1]
				o = o[:len(o)-1]
				i--
			} else if e.msg == msg {
				e.lastSeen = time.Now().Unix()

				if e.amount < 2 {
					e.amount++
					printWithCorrectCaller(msg)
				}

				if e.amount == 2 {
					e.amount++
					log.Println("Muting further repetitive occurrences of this error.")
				}

				o[i] = e
				return
			}
		}

		o = append(o, spamlessError{
			msg:      msg,
			lastSeen: time.Now().Unix(),
		})
	} else {
		spamlessErrors[origin] = []spamlessError{{
			msg:      msg,
			lastSeen: time.Now().Unix(),
		}}
	}

	printWithCorrectCaller(msg)
}

func logInfoIfNoErrorSpam(origin errorOrigin, msg string) {
	if o, ok := spamlessErrors[origin]; ok {
		for _, e := range o {
			if e.amount < 3 && time.Now().Unix()-e.lastSeen < 5*60 {
				return
			}
		}
	}

	printWithCorrectCaller(msg)
}

func resolveErrorSpam(origin errorOrigin) {
	delete(spamlessErrors, origin)
}

func printWithCorrectCaller(msg string) {
	if _, file, line, ok := runtime.Caller(2); ok {
		pathSegments := strings.Split(file, "/")
		log.SetFlags(log.Ldate | log.Ltime)
		log.Printf("%s:%d: %s", pathSegments[len(pathSegments)-1], line, msg)
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		log.Println(msg)
	}
}

type errorOrigin int

const (
	errorOriginInternet errorOrigin = iota
	errorOriginAuth
	errorOriginDefinitions
	errorOriginProfileRequest
	errorOriginActivityPhases
	errorOriginDiscord
	errorOriginSteam
)
