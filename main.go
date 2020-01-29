package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
)

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	kafkaProducer, err :=
		sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	defer kafkaProducer.AsyncClose()

	go ProcessResponse(kafkaProducer)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_SERVER"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := redisClient.Ping().Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", pong)
	defer redisClient.Close()

	c := Controller{kafkaProducer, redisClient}

	httpURI := os.Getenv("HTTP_SERVER_URI")
	httpPath := os.Getenv("HTTP_SERVER_PATH")

	http.HandleFunc(httpPath, c.Handler)
	fmt.Printf("Listening on %s\n", httpURI)
	panic(http.ListenAndServe(httpURI, nil))
}
