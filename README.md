# This project is about to transfer data from HTTP requests into Apache Kafka topics according to pre-defined fields.

## General Flow
- Client sends http request.
- Server reads http request's body.
- Server unmarshals body's json into struct:
    - metadata1 as Topic
    - metadata2 as Token
- Server verifys topic's token with redis key-value store.
- If the token is valid, the server produces request's body into the topic was specified in metadata1. 

## Prerequisites
- Kafka cluster/single broker.
- Redis database.

## Configuration (github.com/spf13/viper)
Teleport can be configured via environment variables starts with TELEPORT.

## Components
- HTTP server (net/http) listens on specific URI

    - URI:                  TELEPORT_HTTP_SERVER_URI=localhost:3333
    - Path:                 TELEPORT_HTTP_SERVER_PATH=/receive

- Kafka producer (github.com/Shopify/sarama) produces requests into kafka topics

    - Brokers:              TELEPORT_KAFKA_BROKERS=broker1:9092,broker2:9092,broker3:9092
    - Use TLS:              TELEPORT_KAFKA_TLS_ENABLED=true
    - Client certificate:   TELEPORT_KAFKA_TLS_CLIENT_CERT=client-cert.pem
    - Client key:           TELEPORT_KAFKA_TLS_CLIENT_KEY=client-key.pem
    - CA certificate:       TELEPORT_KAFKA_TLS_CA_CERT=ca-cert.pem

- Redis client (github.com/go-redis/redis) verifys topics' token

    - Server:               TELEPORT_REDIS_SERVER=redis-server:6379
    - Password:             TELEPORT_REDIS_PASSWORD=mypass
    - DB:                   TELEPORT_REDIS_DB=0

- Logger writes logs into different log files
    - Log directory:        TELEPORT_LOGDIR=/tmp/mylogs


# Links:
- https://ednsquare.com/story/how-to-create-async-producers-with-kafka-go-sarama------RQKdXV
- https://developpaper.com/how-golang-uses-sarama-to-access-kafka/
- https://gobyexample.com/
- https://tour.golang.org/
- https://golangbot.com/goroutines/

# Modules:
- online pre-white: go mod vendor

# Compile:
- online:           make
- offline:          go build -mod vendor

# Run:
- interactive:      run.sh