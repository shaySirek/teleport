package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"
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
	logfileAdmin     = "admin"
	timeFormat       = "2006-01-02 15:04:05.000"
)

type logEvent struct {
	Timestamp string      `json:"timestamp"`
	Host      string      `json:"host"`
	Loglevel  string      `json:"loglevel"`
	Component string      `json:"component"`
	Event     interface{} `json:"event"`
}

// KafkaMessage represents kafka message was produced into a kafka topic
type KafkaMessage struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
}

// ErrorLog represents an error log
type ErrorLog struct {
	Description string `json:"description"`
	Message     string `json:"message"`
}

// WriteLog write log into log file
func WriteLog(logfile string, loglevel string, component string, event interface{}) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	logEventBytes, err := json.Marshal(logEvent{time.Now().Format(timeFormat), hostname, loglevel, component, event})
	if err != nil {
		logEventBytes = []byte{}
	}
	logEvent := string(logEventBytes)

	f, err := os.OpenFile(fmt.Sprintf("%s.log", path.Join(viper.GetString(configLogDir), logfile)),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Println(err.Error())
		log.Println(logEvent)
	}
	defer f.Close()

	logger := log.New(f, "", 0)

	logger.Println(logEvent)
}
