//go:generate mockgen -source email_rabbitmq.go -destination mock/email_rabbitmq.go -package mock
package email

// Emails Publisher interface
type EmailsPublisher interface {
	Publish(body []byte, contentType string) error
}

// Emails Consumer interface
type EmailsConsumer interface {
	Consume(workerPoolSize int, exchange, queueName, bindingKey, consumerTag string) error
}
