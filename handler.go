package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
)

// Controller allows us to attach a kafka producer
// to our handlers
type Controller struct {
	kafkaProducer sarama.AsyncProducer
	redisClient   *redis.Client
}

// TeleportMessage represents http request body
type TeleportMessage struct {
	FileName  string `json:"file_name"`
	Data      string `json:"data"`
	User      string `json:"user"`
	Index     string `json:"index"`
	Parts     int    `json:"parts"`
	ID        string `json:"id"`
	Topic     string `json:"metadata1"`
	Token     string `json:"metadata2"`
	Metadata3 string `json:"metadata3"`
	Metadata4 string `json:"metadata4"`
	Metadata5 string `json:"metadata5"`
	Metadata6 string `json:"metadata6"`
}

// Handler grabs a message http requests,
// looks for topic's token in redis;
// If token is valid, sends it to the kafka topic asynchronously.
func (c *Controller) Handler(w http.ResponseWriter, r *http.Request) {
	body, ioErr := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if ioErr != nil {
		http.Error(w, ioErr.Error(), http.StatusBadRequest)
		return
	}

	var msg TeleportMessage

	if err := json.Unmarshal(body, &msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Printf("%s cannot be decoded\n", body)
		return
	}

	token, redisErr := c.redisClient.Get(msg.Topic).Result()
	if redisErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Printf("%v", redisErr)
		return
	} else if redisErr == redis.Nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Printf("Topic %s does not exist (redis token is required)\n", msg.Topic)
		return
	} else if token != msg.Token {
		w.WriteHeader(http.StatusForbidden)
		fmt.Printf("token \"%s\" is invalid for topic \"%s\"\n", msg.Token, msg.Topic)
		return
	}

	c.kafkaProducer.Input() <- &sarama.ProducerMessage{Topic: msg.Topic, Key: nil, Value: sarama.StringEncoder(body)}
	w.WriteHeader(http.StatusOK)
}
