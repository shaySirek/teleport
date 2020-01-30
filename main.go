package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

const (
	configPrefix                = "teleport"
	configKafkaBrokers          = "kafka_brokers"
	configKafkaBrokersDelimiter = ","
	configRedisServer           = "redis_server"
	configRedisPassword         = "redis_password"
	configRedisDataBase         = "redis_db"
	configHTTPURI               = "http_server_uri"
	configHTTPPath              = "http_server_path"
)

func main() {

	viper.SetDefault(configRedisPassword, "")
	viper.SetDefault(configRedisDataBase, 0)

	viper.SetEnvPrefix(configPrefix)
	viper.AutomaticEnv()
	viper.GetViper().AllowEmptyEnv(true)

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	brokers := strings.Split(viper.Get(configKafkaBrokers).(string), configKafkaBrokersDelimiter)

	kafkaProducer, kafkaErr := sarama.NewAsyncProducer(brokers, config)
	if kafkaErr != nil {
		panic(kafkaErr)
	}
	defer kafkaProducer.AsyncClose()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.Get(configRedisServer).(string),
		Password: viper.Get(configRedisPassword).(string),
		DB:       viper.Get(configRedisDataBase).(int),
	})

	pong, redisErr := redisClient.Ping().Result()
	if redisErr != nil {
		panic(redisErr)
	}
	log.Printf("redis: %s\n", pong)
	defer redisClient.Close()

	c := Controller{kafkaProducer, redisClient}

	httpURI := viper.Get(configHTTPURI).(string)
	httpPath := viper.Get(configHTTPPath).(string)

	http.HandleFunc(httpPath, c.Handler)

	go c.ProcessResponse()

	log.Printf("Listening on %s\n", httpURI)
	httpErr := http.ListenAndServe(httpURI, nil)
	if httpErr != nil {
		panic(httpErr)
	}
}
