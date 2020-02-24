package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
)

// Controller allows us to attach kafka producer and redis client to our HTTP handlers
type Controller struct {
	kafkaProducer sarama.AsyncProducer
	redisClient   *redis.Client
}

// TeleportMessage represents a HTTP request body
type TeleportMessage struct {
	Topic string `json:"metadata1"`
	Token string `json:"metadata2"`
}

// ProcessResponse grabs results and errors from kafka async producer
func (c *Controller) ProcessResponse() {
	for {
		select {
		// Produce was done successfully
		case result := <-c.kafkaProducer.Successes():
			key, kErr := result.Key.Encode()
			value, vErr := result.Value.Encode()
			if kErr != nil {
				key = []byte{}
			}
			if vErr != nil {
				value = []byte{}
			}
			WriteLog(result.Topic, logLevelInfo, componentKafka,
				KafkaMessage{string(key), string(value), result.Topic, result.Partition, result.Offset})
		// Produce was failed
		case err := <-c.kafkaProducer.Errors():
			WriteLog(logfileAdmin, logLevelError, componentKafka, ErrorLog{"Failed to produce message", err.Error()})
		}
	}
}

// Handler grabs http requests,
// looks for topic's token in redis;
// If token is valid, sends body into kafka topic asynchronously.
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	// Read http request body
	body, ioErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if ioErr != nil {
		http.Error(w, ioErr.Error(), http.StatusBadRequest)
		WriteLog(logfileAdmin, logLevelError, componentDecoder, ErrorLog{"Request body cannot be read", ioErr.Error()})
		return
	}

	// Parse body into TeleportMessage struct
	var msg TeleportMessage

	if err := json.Unmarshal(body, &msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		WriteLog(logfileAdmin, logLevelError, componentDecoder, ErrorLog{"Request body cannot be decoded", string(body)})
		return
	}

	// Check if topic exists in redis
	token, redisErr := c.redisClient.Get(msg.Topic).Result()
	if redisErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		WriteLog(logfileAdmin, logLevelError, componentRedis, ErrorLog{fmt.Sprintf("topic %s does not exist in redis", msg.Topic), redisErr.Error()})
		return
	}

	// Check if topic's token is valid
	if token != msg.Token {
		w.WriteHeader(http.StatusForbidden)
		WriteLog(msg.Topic, logLevelError, componentAuth, fmt.Sprintf("token %s is invalid for topic %s", msg.Token, msg.Topic))
		return
	}

	// Produce body as message into kafka topic
	c.kafkaProducer.Input() <- &sarama.ProducerMessage{
		Topic: msg.Topic,
		Key:   sarama.StringEncoder(msg.Topic),
		Value: sarama.StringEncoder(body)}

	// Write 200 OK HTTP response
	w.WriteHeader(http.StatusOK)
}
