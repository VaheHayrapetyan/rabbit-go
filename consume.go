package rabbit_go

import (
	"errors"
	"fmt"
	"github.com/kataras/golog"
	"github.com/streadway/amqp"
	"time"
)

type Message struct {
	amqp.Delivery
}

func (m Message) Body() []byte {
	return m.Delivery.Body
}

type MessageHandler func(message Message)

func (c *client) Consume(consumerName Name, handler MessageHandler) error {
	c.consumeMutex.Lock()
	defer c.consumeMutex.Unlock()
	if !c.opened {
		return errors.New("rabbit error: consume error: connection not opened")
	}
	
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	
	consumer, exists := c.consumers[consumerName]
	if false == exists {
		return errors.New(fmt.Sprintf("rabbit error: consume error: consumer name %s not found", consumerName))
	}

	msg, err := ch.Consume(
		consumer.Queue,     // queue
		consumerName,       // consumer
		consumer.AutoAck,   // auto-ack
		consumer.Exclusive, // exclusive
		consumer.NoLocal,   // no-local
		consumer.NoWait,    // no-wait
		consumer.Args,      // args
	)
	if err != nil {
		return err
	}
	followMessages(msg, consumer.Workers, handler)
	return nil
}

func followMessages(message <-chan amqp.Delivery, worker int, handler MessageHandler) {
	for i := 0; i < worker; i++ {
		go func() {
			defer golog.Info("Connection refused: exited loop")
			for d := range message {
				handler(Message{d})
			}
		}()
	}
}

func testCloseCh(ch *amqp.Channel) {
	go func() {
		t := time.Now()
		golog.Info("start")
		time.Sleep(10 * time.Second)
		golog.Info("done: ", time.Since(t))
		ch.Close()
	}()
}
