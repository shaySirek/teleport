package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ZachtimusPrime/Go-Splunk-HTTP/splunk"
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
	configSplunkURL             = "splunk_url"
	splunkEndpoint              = "https://%s:8088/services/collector"
	configSplunkToken           = "splunk_token"
	configSplunkSource          = "splunk_source"
	configSplunkSourcetype      = "splunk_sourcetype"
	configSplunkIndex           = "splunk_index"
	configSplunkTLSInsecure     = "splunk_tls_insecure"
)

func main() {

	viper.SetDefault(configRedisPassword, "")
	viper.SetDefault(configRedisDataBase, 0)

	viper.SetEnvPrefix(configPrefix)
	viper.AutomaticEnv()
	viper.GetViper().AllowEmptyEnv(true)

	trLoggerhttpClient := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool(configSplunkTLSInsecure)}}
	loggerhttpClient := &http.Client{Timeout: time.Second * 20, Transport: trLoggerhttpClient}
	logger := logger{splunk.NewClient(
		loggerhttpClient,
		fmt.Sprintf(splunkEndpoint, viper.GetString(configSplunkURL)),
		viper.GetString(configSplunkToken),
		viper.GetString(configSplunkSource),
		viper.GetString(configSplunkSourcetype),
		viper.GetString(configSplunkIndex),
	)}

	kafkaProducer, kafkaErr := getKafkaProducer(
		strings.Split(viper.GetString(configKafkaBrokers), configKafkaBrokersDelimiter),
		viper.GetBool(configKafkaTLSEnabled),
		viper.GetString(configKafkaTLSClientCert),
		viper.GetString(configKafkaTLSClientKey),
		viper.GetString(configKafkaTLSCACert),
	)

	if kafkaErr != nil {
		logger.SendLog(logLevelPanic, componentKafka, kafkaErr.Error())
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
		logger.SendLog(logLevelPanic, componentRedis, redisErr.Error())
		panic(redisErr)
	}
	defer redisClient.Close()

	c := Controller{kafkaProducer, redisClient, &logger}

	httpURI := viper.GetString(configHTTPURI)
	httpPath := viper.GetString(configHTTPPath)

	http.HandleFunc(httpPath, c.Handler)
	logger.SendLog(logLevelInfo, componentHTTP, fmt.Sprintf("Listening on %s", httpURI))

	go c.ProcessResponse()

	httpErr := http.ListenAndServe(httpURI, nil)
	if httpErr != nil {
		logger.SendLog(logLevelPanic, componentHTTP, httpErr.Error())
		panic(httpErr)
	}
}
