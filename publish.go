package rabbit_go

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
)

func (c *client) Publish(producerName Name, message []byte) error {
	c.publishMutex.Lock()
	defer c.publishMutex.Unlock()
	
	if !c.opened {
		return errors.New("rabbit error: publish error: connection not opened")
	}
	
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	
	p, exists := c.producers[producerName]
	
	if !exists {
		return errors.New(fmt.Sprintf("rabbit error: publish error: producer %s not found", producerName))
	}
	p.Options.Body = message
	if err := ch.Publish(
		p.Exchange,                 // exchange
		p.RoutingKey,               // routing key
		p.Mandatory,                // mandatory
		p.Immediate,                // immediate
		amqp.Publishing(p.Options), //msg
	); err != nil {
		return err
	}
	
	return nil
}
