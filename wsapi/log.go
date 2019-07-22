// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package wsapi

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// setup subsystem loggers
var (
	rpcLog    *log.Entry
	serverLog *log.Entry
	wsLog     *log.Entry
)

// NewLogFromConfig outputs logs to a file given by logpath
func NewLogFromConfig(logPath, logLevel, prefix string) *log.Entry {
	var logFile io.Writer
	if logPath == "stdout" {
		logFile = os.Stdout
	} else {
		logFile, _ = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
	}

	logger := log.New()
	logger.SetOutput(logFile)
	if strings.ToLower(logLevel) == "none" {
		log.SetLevel(log.PanicLevel)
		log.SetOutput(ioutil.Discard)
	} else {
		lvl, err := log.ParseLevel(logLevel)
		if err != nil {
			panic(err)
		}
		logger.SetLevel(lvl)
	}
	return logger.WithField("prefix", prefix)
}

func InitLogs(logPath, logLevel string) {
	rpcLog = NewLogFromConfig(logPath, logLevel, "RPC")
	serverLog = NewLogFromConfig(logPath, logLevel, "SERV")
	wsLog = NewLogFromConfig(logPath, logLevel, "WSAPI")
}
