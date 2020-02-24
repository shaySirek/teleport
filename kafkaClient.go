package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

	"github.com/Shopify/sarama"
)

// Generate tls configuration from certificates files
func getTLSConfig(clientcertfile, clientkeyfile, cacertfile string) (*tls.Config, error) {
	// Load client cert
	clientcert, err := tls.LoadX509KeyPair(clientcertfile, clientkeyfile)
	if err != nil {
		return nil, err
	}

	// Load CA cert pool
	cacert, err := ioutil.ReadFile(cacertfile)
	if err != nil {
		return nil, err
	}
	cacertpool := x509.NewCertPool()
	cacertpool.AppendCertsFromPEM(cacert)

	// Generate tls config
	tlsConfig := tls.Config{}
	tlsConfig.RootCAs = cacertpool
	tlsConfig.Certificates = []tls.Certificate{clientcert}
	tlsConfig.BuildNameToCertificate()

	return &tlsConfig, nil
}

// Create kafka producer client
func getKafkaProducer(brokers []string, tlsEnabled bool, tlsClientCert string, tlsClientKey string, tlsCACert string) (sarama.AsyncProducer, error) {
	// Create kafka producer config
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	// Configure tls if it's required
	if tlsEnabled {
		tlsConfig, err := getTLSConfig(tlsClientCert, tlsClientKey, tlsCACert)
		if err != nil {
			log.Fatal(err)
		}

		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	return sarama.NewAsyncProducer(brokers, config)
}
