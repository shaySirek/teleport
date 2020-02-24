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
	// Set default configurations
	viper.SetDefault(configRedisPassword, "")
	viper.SetDefault(configRedisDataBase, 0)

	// Set configuration auto prefix
	viper.SetEnvPrefix(configPrefix)
	viper.AutomaticEnv()
	viper.GetViper().AllowEmptyEnv(true)

	// Create kafka producer client according to relevant configuration
	kafkaProducer, kafkaErr := getKafkaProducer(
		strings.Split(viper.GetString(configKafkaBrokers), configKafkaBrokersDelimiter),
		viper.GetBool(configKafkaTLSEnabled),
		viper.GetString(configKafkaTLSClientCert),
		viper.GetString(configKafkaTLSClientKey),
		viper.GetString(configKafkaTLSCACert),
	)

	// Handle error in creating kafka producer
	if kafkaErr != nil {
		WriteLog(logfileAdmin, logLevelPanic, componentKafka, kafkaErr.Error())
		panic(kafkaErr)
	}
	defer kafkaProducer.AsyncClose()

	// Create redis client according to relevant configuration
	redisClient := redis.NewClient(&redis.Options{
		Addr:     viper.GetString(configRedisServer),
		Password: viper.GetString(configRedisPassword),
		DB:       viper.GetInt(configRedisDataBase),
	})

	// Check redis connection health
	redisErr := redisClient.Ping().Err()
	if redisErr != nil {
		WriteLog(logfileAdmin, logLevelPanic, componentRedis, redisErr.Error())
		panic(redisErr)
	}
	defer redisClient.Close()

	// Create Controller instance
	c := Controller{kafkaProducer, redisClient}

	// Configure HTTP server
	httpURI := viper.GetString(configHTTPURI)
	httpPath := viper.GetString(configHTTPPath)

	http.HandleFunc(httpPath, c.Handler)
	WriteLog(logfileAdmin, logLevelInfo, componentHTTP, fmt.Sprintf("Listening on %s", httpURI))

	// Start ProcessResponse as go routine
	go c.ProcessResponse()

	// Start HTTP server and handle error
	httpErr := http.ListenAndServe(httpURI, nil)
	if httpErr != nil {
		WriteLog(logfileAdmin, logLevelPanic, componentHTTP, httpErr.Error())
		panic(httpErr)
	}
}
