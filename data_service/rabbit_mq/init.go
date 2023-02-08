package rabbit_mq

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	DialError       = errors.New("dial MQ err")
	ConnectionError = errors.New("connection nil err")
	ChannelGetError = errors.New("channel get err")
	RabbitMQInitErr = []error{DialError}
)

type RabbitMQConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

func InitMQ(ctx context.Context, ch *amqp.Channel, config *RabbitMQConfig) error {
	var (
		conn *amqp.Connection
		err  error
		addr string
	)

	addr = fmt.Sprintf("amqp://%s:%s@%s:%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
	)
	logrus.Infof("Init MQ  [addr:%s]", addr)

	if conn, err = amqp.DialConfig(addr, amqp.Config{Vhost: "/"}); err != nil {
		return errors.Join(DialError, err)
	}
	if conn == nil {
		return errors.Join(ConnectionError, err)
	}
	go func() {
		<-ctx.Done()
		if conn != nil {
			conn.Close()
		}
	}()

	if ch, err = conn.Channel(); err != nil {
		return errors.Join(ChannelGetError, err)
	}

	logrus.Infof("Init MQ  [addr:%s] success", addr)

	return nil
}

var (
	DeclareQueueErr    = errors.New("declare queue err")
	DeclareQueueErrArr = []error{DeclareQueueErr}
)

func DeclareQueue(ch *amqp.Channel, queue *amqp.Queue, name string, isPriority bool, maxPriority uint8) error {
	var (
		err  error
		args = make(amqp.Table)
	)
	if isPriority {
		args["x-max-priority"] = maxPriority
	}
	if *queue, err = ch.QueueDeclare(name, true, false, false, false, args); err != nil {
		return err
	}
	//todo success

	return err
}
