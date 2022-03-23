package utils

import (
	log "github.com/sirupsen/logrus"
	"runtime/debug"
)

func HandleError() {
	if err := recover(); err != nil {
		log.Errorln(err)
		debug.PrintStack()
	}
}
