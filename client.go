package rabbit_go

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"os"
	"sync"
)

func newClient() *client {
	return &client{
		conn:         nil,
		opened:       false,
		queues:       nil,
		producers:    nil,
		consumers:    nil,
		mutex:        sync.Mutex{},
		declareMutex: sync.Mutex{},
		consumeMutex: sync.Mutex{},
		publishMutex: sync.Mutex{},
	}
}

func NewClient(data []byte) (Client, error) {
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	config.adapt()

	c := newClient()
	c.queues = config.Queues
	c.consumers = config.Consumers
	c.producers = config.Producers

	return c, nil
}

func NewClientFromFile(jsonQCPPath string) (Client, error) {
	c := newClient()
	err := c.initWithConfigFile(jsonQCPPath)
	return c, err
}

func (c *client) initWithConfigFile(jsonConfigFilePath string) error {
	data, err := os.ReadFile(jsonConfigFilePath)
	if err != nil {
		return err
	}

	var config Config

	if err = json.Unmarshal(data, &config); err != nil {
		return err
	}

	config.adapt()

	c.queues = config.Queues
	c.consumers = config.Consumers
	c.producers = config.Producers

	return nil
}

func (c *client) OpenConn(url string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.opened {
		return errors.New("rabbit error: open error: connection already opened")
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}

	c.conn = conn
	c.opened = true

	err = c.declareAllQueues()
	if nil != err {
		c.conn.Close()
		c.opened = false

		return err
	}
	return nil
}

func (c *client) CloseConn() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.opened {
		return errors.New("rabbit error: close error: connection not opened")
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	c.opened = false
	return nil
}

func (c *client) IsOpen() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.opened
}

func (c *client) declareAllQueues() error {
	for name := range c.queues {
		err := c.declareQueue(name)
		if nil != err {
			return err
		}
	}
	return nil
}

func (c *client) declareQueue(name Name) error {
	c.declareMutex.Lock()
	defer c.declareMutex.Unlock()
	if !c.opened {
		return errors.New("rabbit error: declare error: connection not opened")
	}

	q, exists := c.queues[name]
	if !exists {
		return errors.New(fmt.Sprintf("rabbit error: declare error: queue %s not found", name))
	}

	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if _, err := ch.QueueDeclare(
		name,                 // name
		q.Options.Durable,    // durable
		q.Options.AutoDelete, // delete when unused
		q.Options.Exclusive,  // exclusive
		q.Options.NoWait,     // no-wait
		q.Options.Args,       // arguments
	); err != nil {
		return err
	}

	//err = ch.Qos(
	//	1000, // prefetch count
	//	0,     // prefetch size
	//	false, // global
	//)

	if err := ch.QueueBind(
		name,
		q.RoutingKey,
		q.Exchange,
		q.Options.NoWait,
		q.Options.Args,
	); err != nil {
		return err
	}
	return nil
}
