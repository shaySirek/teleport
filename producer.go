package main

import (
	"log"

	"github.com/Shopify/sarama"
)

// ProcessResponse grabs results and errors from a producer
// asynchronously
func ProcessResponse(producer sarama.AsyncProducer) {
	for {
		select {
		case result := <-producer.Successes():
			log.Printf("%v message: \"%s\" was sent to topic %s partition %d at offset %d\n", result.Timestamp.Format(""), result.Value, result.Topic, result.Partition, result.Offset)
		case err := <-producer.Errors():
			log.Println("Failed to produce message", err)
		}
	}
}
