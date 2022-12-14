package rabbitmq

import (
	"context"
	"log"
	"rmq_service/internal/email"
	"rmq_service/pkg/logger"

	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
)

const (
	exchangeKind 				= "direct"
	exchangeDurable  		= true
	exchangeAutoDelete 	= false
	exchangeInternal 		= false
	exchangeNoWait 			= false

	queueDurable 				= true
	queueAutoDelete 		= false
	queueExclusive 			= false
	queueNoWait 				= false

	publishMandatory 		= false
	publishImmediate 		= false

	prefetchCount 			= 1
	prefetchSize 				= 0
	prefetchGlobal 			= false

	consumeAutoAck 			= false
	consumeExclusive 		= false
	consumeNoLocal 			= false
	consumeNoWait 			= false
)

var (
	incomingMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_incoming_rabbitmq_messages_total",
		Help: "The total number of incoming RabbitMQ messages",
	})

	successMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_success_incoming_rabbitmq_messages_total",
		Help: "The total number of success incoming success RabbitMQ messages",
	})

	errorMessages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "emails_error_incoming_rabbitmq_messages_total",
		Help: "The total number of error incoming success RabbitMQ messages",
	})
)

// Images RabbitMQ Consumer
type EmailsConsumer struct {
	amqpConn 	*amqp.Connection
	logger 		logger.Logger
	emailUC		email.EmailsUseCase
}

// Images Consumer constructor
func NewImagesConsumer(
	amqpConn *amqp.Connection,
	logger logger.Logger,
	emailUC email.EmailsUseCase,
) *EmailsConsumer {
	return &EmailsConsumer{amqpConn: amqpConn, logger: logger, emailUC: emailUC}
}

// Creates channel to consume messages
func (c *EmailsConsumer) CreateChannel(
	exchangeName, queueName, bindingKey, consumerTag string,
) (*amqp.Channel, error) {
	ch, err := c.amqpConn.Channel()
	if err != nil {
		log.Fatalf("Consumer::Channel(): %v", err)
		return nil, err
	}
	
	c.logger.Infof("Declaring exchange: %s", exchangeName)
	err = ch.ExchangeDeclare(
		exchangeName,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	)
	if err != nil {
		log.Fatalf("Consumer::ExchangeDeclare(): %v", err)
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil,
	)
	if err != nil {
		log.Fatalf("Consumer::QueueDeclare(): %v", err)
		return nil, err
	}

	c.logger.Infof("Declared queue, binding it to exchange:\nQueue: %v\n" + 
		"Message Count: %v\nConsumer Count: %v\nExchange: %v, Binding Key: %v",
		queue.Name,
		queue.Messages,
		queue.Consumers,
		exchangeName,
		bindingKey,
	)

	err = ch.QueueBind(
		queue.Name,
		bindingKey,
		exchangeName,
		queueNoWait,
		nil,
	)
	if err != nil {
		log.Fatalf("Consumer::QueueBing(): %v", err)
	}

	c.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)

	err = ch.Qos(prefetchCount, prefetchSize, prefetchGlobal)
	if err != nil {
		log.Fatalf("Consumer::Qos(): %v", err)
		return nil, err
	}

	return ch, nil
}

func (c *EmailsConsumer) worker(ctx context.Context, messages <-chan amqp.Delivery) {
	for delivery := range messages {
		span, ctx := opentracing.StartSpanFromContext(ctx, "EmailsConsumer.worker")

		c.logger.Infof("processDeliveries deliveryTag: %v", delivery.DeliveryTag)

		incomingMessages.Inc()

		err := c.emailUC.SendEmails(ctx, delivery.Body)
		if err != nil {
			if err := delivery.Reject(false); err != nil {
				c.logger.Errorf("Error delivery.Reject: %v", err)
			}

			c.logger.Errorf("Failed to process delivery: %v", err)
			errorMessages.Inc()
		} else {
			if err = delivery.Ack(false); err != nil {
				c.logger.Errorf("Failed to acknowledge the message: %v", err)
			}
		}
		span.Finish()
	}

	c.logger.Info("Deliveries channel closed")
}

// Start new rabbitmq consumer
func (c *EmailsConsumer) StartConsumer(
	workerPoolSize int,
	exchange, queueName, bindingKey, consumerTag string,
) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := c.CreateChannel(exchange, queueName, bindingKey, consumerTag)
	defer ch.Close()
	if err != nil {
		log.Fatalf("Consumer::StartConsumer()::CreateChannel(): %v", err)
		return err
	}

	deliveries, err := ch.Consume(
		queueName,
		consumerTag,
		consumeAutoAck,
		consumeExclusive,
		consumeNoLocal,
		consumeNoWait,
		nil,
	)
	if err != nil {
		log.Fatalf("Consumer::StartConsumer()::Consume(): %v", err)
		return err
	}

	for i := 0; i < workerPoolSize; i++ {
		go c.worker(ctx, deliveries)
	}

	chanErr := <-ch.NotifyClose(make(chan *amqp.Error))
	c.logger.Errorf("ch.NotifyClose(): %v", chanErr)
	return chanErr
}
