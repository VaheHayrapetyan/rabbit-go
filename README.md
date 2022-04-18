# rabbit-go

Hello. Here you will get acquainted with the rabbit-go library. The library is designed to work with RabbitMQ at a high level through configuration files.      
Similarly, if you are familiar with RabbitMQ, you are also familiar with the concept of Consumer in the context of RabbitMQ. In the rabbit-go library, you can set the number of workers for each consumer that will handle read messages in parallel.

Functions:

    func NewClient(data []byte) (Client, error)     
    func NewClientFromFile(jsonQCPPath string)  (Client, error)

    func (c Client) OpenConn(url string) error
    func (c Client) CloseConn() error
    func (c Client) IsOpen() bool
    func (c Client) Consume(consumerName Name, handler MessageHandler) error
    func (c Client) Publish(producerName Name, message []byte) error
    func (c Client) Queues() map[Name]Queue
    func (c Client) Producers() map[Name]Producer
    func (c Client) Consumers() map[Name]Consumer

    func (q *Queue) PutArgs(name string, val interface{})
    func (c Consumer) WorkersCount() int

Structures:

    type Name = string

    type Config struct {
        Queues    map[Name]Queue    `json:"queues"`
        Producers map[Name]Producer `json:"producers"`
        Consumers map[Name]Consumer `json:"consumers"`
    }

    type Queue struct {
        Exchange   string  `json:"exchange"`
        RoutingKey string  `json:"routing_key"`
        Options    Options `json:"options"`
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

    type Producer struct {
        Exchange   string     `json:"exchange"`
        RoutingKey string     `json:"routing_key"`
        Mandatory  bool       `json:"mandatory"`
        Immediate  bool       `json:"immediate"`
        Options    Publishing `json:"options"`
    }

    type Options struct {
        Durable    bool       `json:"durable"`
        AutoDelete bool       `json:"auto_delete"`
        Exclusive  bool       `json:"exclusive"`
        NoWait     bool       `json:"no_wait"`
        Args       amqp.Table `json:"args"`
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

    type Message struct {
        amqp.Delivery
    }

Now let's start from the beginning․     
You must use one of these two functions with which you will create the Client. After that, through the configurations you set, you will have queues, producers and consumers.

    func NewClient(data []byte) (Client, error)     
    func NewClientFromFile(jsonQCPPath string)  (Client, error)


You need to pass data to the NewClient function, which is a byte representation of the Config structure in json format.     
The NewClientFromFile function must be passed the path to the json file containing the configurations corresponding to the Config structure.

    //config file example
    {   //there are two queues: "request_q", "response_q"
        "queues": {
            "request_q": { //the queue bound to "ex" exchange with "request" routing key
                "exchange": "ex",
                "routing_key": "request",
                "options": {
                    "durable": false,
                    "auto_delete": true
                }
            },
            "response_q": { //the queue bound to "ex" exchange with "response" routing key
                "exchange": "ex",
                "routing_key": "response",
                "options": {
                    "durable": false,
                    "auto_delete": true
                }
            }
        },
        
        //there are two consumers: "request_c", "response_c"
        "consumers": {
            "request_c": { //the consumer reads messages from "request_q" queue, and handled those messages with 4 workers
                "queue": "request_q",
                "workers": 4
            },
            "response_c": { //the consumer reads messages from "response_q" queue, and handled those messages with 4 workers
                "queue": "response_q",
                "workers": 4
            }
        },
    
        //there is one producer: "request_p"
        "producers": {
            "request_p": { //the producer sends message to exchange "ex" with routing key "request"
                "exchange": "ex",
                "routing_key": "request",
                "options": {
                    "content_type": "application/json",
                    "delivery_mode": 2,
                    "reply_to": "response" //the producer reads the response from the queue, which is bound to the exchange "ex" with the routing key "response"
                }
            }
        }
    }
After using the NewClient or NewClientFromFile functions, you will have access to the Client interface:

    type Client interface {
        OpenConn(url string) error
        CloseConn() error
        IsOpen() bool
        Consume(consumerName Name, handler MessageHandler) error
        Publish(producerName Name, message []byte) error
        Queues() map[Name]Queue
    }

Let's take a look at the purpose of each of these functions.

    func (c Client) OpenConn(url string) error

OpenConn function accepts a string in the AMQP URI format and create a new Connection for your Client. It can return an error for example when the connection is already open.

    func (c Client) CloseConn() error

CloseConn function just closes the connection․ It can return an error for example when the connection not opened.

    func (c Client) IsOpen() bool

IsOpen function returns true if the connection is open on the client and false otherwise.

    func (c Client) Consume(consumerName Name, handler MessageHandler) error

In Consume function you must pass consumer name, and handler function for the messages. To understand the latter, look here։ we have the following in our rabbit_go library

    type Message struct {
    	amqp.Delivery
    }

    type MessageHandler func(message Message)

    func (m Message) Body() []byte

As you remember, in the Consumer structure was also given the name of the queue. By passing the name of the consumer to this function, you can read all messages from this queue and handle them with a function of type MessageHandler.
Consume function can return an error for example when the connection not opened, or when consumer name not found.

    func  (c Client) Publish(producerName Name, message []byte) error

In Publish function you must pass producer name and message. As you remember, in the Producer structure was also given the exchange name and routing key. By passing the name of the producer to this function, you can send the message to that producer's exchange with that roting key.
Publish function can return an error for example when the connection not opened, or when producer name not found.

    func (c Client) Queues() map[Name]Queue

Queues function returns all queues on that client․

Producers and consumers are returned through the Producers and Consumers functions, respectively:

    func (c Client) Producers() map[Name]Producer
    func (c Client) Consumers() map[Name]Consumer

For add arguments on the queues, you must use this function:

    func (q *Queue) PutArgs(name string, val interface{})

As you remember, in the Consumer structure was also given the count of workers, which determines the number of workers or goroutines that will parallel read the messages from queue and work with them.          
To get the count of workers you need to use this function:

    func (c Consumer) WorkersCount() int

That's all !!! Thank you for your attention.