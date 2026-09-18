package platform

type SQSPublisher interface {
	Publish(queueName string, message []byte) error
}

type SQSConsumer interface {
	Consume(queueName string, handler func(message []byte) error) error
}