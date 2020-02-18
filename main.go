package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

const (
	configPrefix                = "teleport"
	configKafkaBrokers          = "kafka_brokers"
	configKafkaBrokersDelimiter = ","
	configKafkaTLSEnabled       = "kafka_tls_enabled"
	configKafkaTLSClientCert    = "kafka_tls_client_cert"
	configKafkaTLSClientKey     = "kafka_tls_client_key"
	configKafkaTLSCACert        = "kafka_tls_ca_cert"
	configRedisServer           = "redis_server"
	configRedisPassword         = "redis_password"
	configRedisDataBase         = "redis_db"
	configHTTPURI               = "http_server_uri"
	configHTTPPath              = "http_server_path"
	configLogDir                = "logdir"
)

func main() {

	viper.SetDefault(configRedisPassword, "")
	viper.SetDefault(configRedisDataBase, 0)

	viper.SetEnvPrefix(configPrefix)
	viper.AutomaticEnv()
	viper.GetViper().AllowEmptyEnv(true)

	kafkaProducer, kafkaErr := getKafkaProducer(
		strings.Split(viper.GetString(configKafkaBrokers), configKafkaBrokersDelimiter),
		viper.GetBool(configKafkaTLSEnabled),
		viper.GetString(configKafkaTLSClientCert),
		viper.GetString(configKafkaTLSClientKey),
		viper.GetString(configKafkaTLSCACert),
	)

	if kafkaErr != nil {
		WriteLog(logfileAdmin, logLevelPanic, componentKafka, kafkaErr.Error())
		panic(kafkaErr)
	}
	defer kafkaProducer.AsyncClose()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString(configRedisServer),
		Password: viper.GetString(configRedisPassword),
		DB:       viper.GetInt(configRedisDataBase),
	})
	redisErr := redisClient.Ping().Err()

	if redisErr != nil {
		WriteLog(logfileAdmin, logLevelPanic, componentRedis, redisErr.Error())
		panic(redisErr)
	}
	defer redisClient.Close()

	c := Controller{kafkaProducer, redisClient}

	httpURI := viper.GetString(configHTTPURI)
	httpPath := viper.GetString(configHTTPPath)

	http.HandleFunc(httpPath, c.Handler)
	WriteLog(logfileAdmin, logLevelInfo, componentHTTP, fmt.Sprintf("Listening on %s", httpURI))

	go c.ProcessResponse()

	httpErr := http.ListenAndServe(httpURI, nil)
	if httpErr != nil {
		WriteLog(logfileAdmin, logLevelPanic, componentHTTP, httpErr.Error())
		panic(httpErr)
	}
}
