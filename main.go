package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
)

const (
	envKafkaBrokers          = "KAFKA_BROKERS"
	envKafkaBrokersDelimiter = ","
	envRedisServer           = "REDIS_SERVER"
	envHTTPURI               = "HTTP_SERVER_URI"
	envHTTPPath              = "HTTP_SERVER_PATH"
)

func main() {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	brokers := strings.Split(os.Getenv(envKafkaBrokers), envKafkaBrokersDelimiter)

	kafkaProducer, kafkaErr :=
		sarama.NewAsyncProducer(brokers, config)
	if kafkaErr != nil {
		panic(kafkaErr)
	}
	defer kafkaProducer.AsyncClose()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv(envRedisServer),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, redisErr := redisClient.Ping().Result()
	if redisErr != nil {
		panic(redisErr)
	}
	fmt.Printf("redis: %s\n", pong)
	defer redisClient.Close()

	c := Controller{kafkaProducer, redisClient}

	httpURI := os.Getenv(envHTTPURI)
	httpPath := os.Getenv(envHTTPPath)

	http.HandleFunc(httpPath, c.Handler)

	go ProcessResponse(kafkaProducer)

	fmt.Printf("Listening on %s\n", httpURI)
	httpErr := http.ListenAndServe(httpURI, nil)
	if httpErr != nil {
		panic(httpErr)
	}
}
