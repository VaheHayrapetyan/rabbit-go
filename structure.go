package rabbit_go

import (
	"github.com/streadway/amqp"
	"math"
	"sync"
	"time"
)

type Config struct {
	Queues    map[Name]Queue    `json:"queues"`
	Producers map[Name]Producer `json:"producers"`
	Consumers map[Name]Consumer `json:"consumers"`
}

// when args type is float but really type is integer adapt function will change that type integer
func (conf *Config) adapt() {
	for k, v := range conf.Queues {
		args := v.Options.Args
		for kArg, vArg := range args {
			switch vArg.(type) {
			case float32, float64:
				fv := vArg.(float64)
				if fv == math.Trunc(fv) {
					conf.Queues[k].Options.Args[kArg] = int64(fv)
				}
			}
		}
	}

	for k, v := range conf.Consumers {
		args := v.Args
		for kArg, vArg := range args {
			switch vArg.(type) {
			case float32, float64:
				fv := vArg.(float64)
				if fv == math.Trunc(fv) {
					conf.Consumers[k].Args[kArg] = int64(fv)
				}
			}
		}
	}

	for k, v := range conf.Producers {
		headers := v.Options.Headers
		for kHeader, vHeader := range headers {
			switch vHeader.(type) {
			case float32, float64:
				fv := vHeader.(float64)
				if fv == math.Trunc(fv) {
					conf.Producers[k].Options.Headers[kHeader] = int64(fv)
				}
			}
		}
	}

}

type Name = string

type Queue struct {
	Exchange   string  `json:"exchange"`
	RoutingKey string  `json:"routing_key"`
	Options    Options `json:"options"`
}

// PutArgs a function that is not needed
func (q *Queue) PutArgs(name string, val interface{}) {
	if nil == q.Options.Args {
		q.Options.Args = amqp.Table{}
	}
	q.Options.Args[name] = val
}

type Options struct {
	Durable    bool       `json:"durable"`
	AutoDelete bool       `json:"auto_delete"`
	Exclusive  bool       `json:"exclusive"`
	NoWait     bool       `json:"no_wait"`
	Args       amqp.Table `json:"args"`
}

type Consumer struct {
	AutoAck   bool       `json:"auto_ack"`
	Exclusive bool       `json:"exclusive"`
	NoLocal   bool       `json:"no_local"`
	NoWait    bool       `json:"no_wait"`
	Args      amqp.Table `json:"args"`
	Queue     Name       `json:"queue"`
	Workers   int        `json:"workers"`
}

// WorkersCount
// Return workers count if 0 will return 1
func (c Consumer) WorkersCount() int {
	if 0 == c.Workers {
		return 1
	}
	return c.Workers
}

type Producer struct {
	Exchange   string     `json:"exchange"`
	RoutingKey string     `json:"routing_key"`
	Mandatory  bool       `json:"mandatory"`
	Immediate  bool       `json:"immediate"`
	Options    Publishing `json:"options"`
}

type Client interface {
	OpenConn(url string) error
	CloseConn() error
	IsOpen() bool
	Consume(consumerName Name, handler MessageHandler) error
	Publish(producerName Name, message []byte) error
	Queues() map[Name]Queue
}
type client struct {
	conn         *amqp.Connection
	opened       bool
	queues       map[Name]Queue
	producers    map[Name]Producer
	consumers    map[Name]Consumer
	mutex        sync.Mutex
	declareMutex sync.Mutex
	consumeMutex sync.Mutex
	publishMutex sync.Mutex
}

func (c *client) Queues() map[Name]Queue {
	return c.queues
}

func (c *client) Producers() map[Name]Producer {
	return c.producers
}

func (c *client) Consumers() map[Name]Consumer {
	return c.consumers
}

type Publishing struct {
	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	Headers amqp.Table `json:"headers"`

	// Properties
	ContentType     string    `json:"content_type"`     // MIME content type
	ContentEncoding string    `json:"content_encoding"` // MIME content encoding
	DeliveryMode    uint8     `json:"delivery_mode"`    // Transient (0 or 1) or Persistent (2)
	Priority        uint8     `json:"priority"`         // 0 to 9
	CorrelationId   string    `json:"correlation_id"`   // correlation identifier
	ReplyTo         string    `json:"reply_to"`         // address to to reply to (ex: RPC)
	Expiration      string    `json:"expiration"`       // message expiration spec
	MessageId       string    `json:"message_id"`       // message identifier
	Timestamp       time.Time `json:"timestamp"`        // message timestamp
	Type            string    `json:"type"`             // message type name
	UserId          string    `json:"user_id"`          // creating user id - ex: "guest"
	AppId           string    `json:"app_id"`           // creating application id

	// The application specific payload of the message
	Body []byte `json:"body"`
}
