package main

import (
	"fmt"
	"github.com/ZachtimusPrime/Go-Splunk-HTTP/splunk"
)

const (
	logLevelInfo     = "INFO"
	logLevelError    = "ERROR"
	logLevelPanic    = "PANIC"
	componentKafka   = "Kafka Producer"
	componentRedis   = "Redis Client"
	componentHTTP    = "HTTP Server"
	componentDecoder = "Request Decoder"
	componentAuth    = "Authentication"
)

type logger struct {
	splunkClient *splunk.Client
}

type logEvent struct {
	Loglevel  string
	Component string
	Msg       interface{}
}

// SendLog sends log to splunk
func (logger *logger) SendLog(loglevel string, component string, msg interface{}) error {
	if err := logger.splunkClient.Log(logEvent{loglevel, component, msg}); err != nil {
		fmt.Printf("%s", err.Error())
		return err
	}
	return nil
}
