package iamqp

import (
	"sync/atomic"
	"time"

	"github.com/streadway/amqp"
)

// channel wrap amqp.Channel with channel ErrClose handled
type channel struct {
	conn          *amqp.Connection
	ch            *amqp.Channel
	pool          chan *channel
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	done          chan bool
	reliable      bool
	available     atomic.Value
}

// handleReconnect will wait for a channel error
// and then continuously attempt to re-initialize
func (c *channel) handleReInit(connection *amqp.Connection) {
	defer func() {
		c.available.Store(false)
	}()

	for {
		c.available.Store(true)
		select {
		case <-c.done:
			Log("iamqp: publish channel done signal received. exit")
			return
		case err, ok := <-c.notifyClose:
			c.available.Store(false)
			if !ok {
				<-time.After(reInitDelay)
				Log("iamqp: reopen...")
			}
			Log("iamqp: channel closed. reInit... err: %v", err)
		}
		err := c.initPublishChannel()
		if err != nil {
			c.available.Store(false)
			Log("iamqp: channel init fail, exit. err:%v", err)
			return
		}
	}
}

func newChannel(session *Session) (*channel, error) {
	c := &channel{
		conn:     session.conn,
		pool:     session.publishChannel,
		reliable: session.reliable,
		done:     make(chan bool, 1),
	}
	c.notifyClose = make(chan *amqp.Error, 1)
	if c.reliable {
		c.notifyConfirm = make(chan amqp.Confirmation, 1)
	}
	err := c.initPublishChannel()
	if err != nil {
		return nil, err
	}
	go c.handleReInit(session.conn)
	return c, nil
}

func (c *channel) initPublishChannel() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	if c.reliable {
		err = c.ch.Confirm(false)
		if err != nil {
			Log("iamqp: should not into here: reopen channel set confirm mode fail.")
			return err
		}
		c.notifyConfirm = c.ch.NotifyPublish(c.notifyConfirm)
	}
	c.notifyClose = channel.NotifyClose(make(chan *amqp.Error, 1))
	c.ch = channel
	return nil
}

func (c *channel) initConsumeChannel() error {
	channel, err := c.conn.Channel()
	if err != nil {
		return err
	}
	c.notifyClose = channel.NotifyClose(make(chan *amqp.Error, 1))
	c.ch = channel
	return nil
}

func newConsumerChannel(session *Session) (*channel, error) {
	c := &channel{
		conn:     session.conn,
		reliable: false,
		done:     make(chan bool, 1),
	}

	err := c.initConsumeChannel()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// Done stop channel reInit
func (c *channel) Done() {
	c.done <- true
}

// Close return channel to pool
func (c *channel) Close() {
	if c.available.Load().(bool) {
		c.pool <- c
	}
}

// publish wrap amqp channel.publish
func (c *channel) publish(exchange, key string, msg Message, opt ...PubOpt) error {
	option := &PublishOption{
		mandatory: false,
		immediate: false,
	}
	for _, o := range opt {
		o(option)
	}
	publishing := amqp.Publishing{
		Headers:         msg.Header,
		Body:            msg.Body,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		Timestamp:       time.Now(),
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationID,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageID,
		Type:            msg.Type,
		UserId:          msg.UserID,
		AppId:           msg.AppID,
	}
	return c.ch.Publish(exchange, key, option.mandatory, option.immediate, publishing)
}

// PublishOption ...
type PublishOption struct {
	mandatory bool
	immediate bool
}

// PubOpt ...
type PubOpt func(opt *PublishOption)

// SetMandatory ...
func SetMandatory(mandatory bool) PubOpt {
	return func(opt *PublishOption) {
		opt.mandatory = mandatory
	}
}

// SetImmediate ...
func SetImmediate(immediate bool) PubOpt {
	return func(opt *PublishOption) {
		opt.immediate = immediate
	}
}

// Message wrapped Publishing and Delivery
type Message struct {
	Header map[string]interface{}

	// Properties
	ContentType     string    // MIME content type
	ContentEncoding string    // MIME content encoding
	DeliveryMode    uint8     // Transient (0 or 1) or Persistent (2)
	Priority        uint8     // 0 to 9
	CorrelationID   string    // correlation identifier
	ReplyTo         string    // address to to reply to (ex: RPC)
	Expiration      string    // message expiration spec
	MessageID       string    // message identifier
	Timestamp       time.Time // message timestamp
	Type            string    // message type name
	UserID          string    // creating user id - ex: "guest"
	AppID           string    // creating application id

	Acknowledger amqp.Acknowledger // the channel from which this delivery arrived
	// Valid only with channel.Consume
	ConsumerTag string

	// Valid only with channel.Get
	MessageCount uint32

	DeliveryTag uint64
	Redelivered bool
	Exchange    string // basic.publish exchange
	RoutingKey  string // basic.publish routing key

	Body []byte
}

func message(delivery amqp.Delivery) *Message {
	return &Message{
		Header: delivery.Headers,

		// Properties
		ContentType:     delivery.ContentType,
		ContentEncoding: delivery.ContentEncoding,
		DeliveryMode:    delivery.DeliveryMode,
		Priority:        delivery.Priority,
		CorrelationID:   delivery.CorrelationId,
		ReplyTo:         delivery.ReplyTo,
		Expiration:      delivery.Expiration,
		MessageID:       delivery.MessageId,
		Timestamp:       delivery.Timestamp,
		Type:            delivery.Type,
		UserID:          delivery.UserId,
		AppID:           delivery.AppId,

		Acknowledger: delivery.Acknowledger, // the channel from which this delivery arrived
		// Valid only with channel.Consume
		ConsumerTag: delivery.ConsumerTag,

		// Valid only with channel.Get
		MessageCount: delivery.MessageCount,

		DeliveryTag: delivery.DeliveryTag,
		Redelivered: delivery.Redelivered,
		Exchange:    delivery.Exchange,
		RoutingKey:  delivery.RoutingKey,

		Body: delivery.Body,
	}
}

/*
Ack delegates an acknowledgement through the Acknowledger interface that the
client or server has finished work on a delivery.
*/
func (m *Message) Ack(multiple bool) error {
	if m.Acknowledger == nil {
		return errDeliveryNotInitialized
	}
	return m.Acknowledger.Ack(m.DeliveryTag, multiple)
}

/*
Reject delegates a negatively acknowledgement through the Acknowledger interface.
*/
func (m *Message) Reject(requeue bool) error {
	if m.Acknowledger == nil {
		return errDeliveryNotInitialized
	}
	return m.Acknowledger.Reject(m.DeliveryTag, requeue)
}

/*
Nack negatively acknowledge the delivery of message(s) identified by the
delivery tag from either the client or server.
*/
func (m *Message) Nack(multiple, requeue bool) error {
	if m.Acknowledger == nil {
		return errDeliveryNotInitialized
	}
	return m.Acknowledger.Nack(m.DeliveryTag, multiple, requeue)
}

// consume
func (c *channel) consume(queue string, option *ConsumeOption) (<-chan amqp.Delivery, error) {
	delivery, err := c.ch.Consume(queue, option.Consumer, option.AutoAck, option.Exclusive,
		true, option.NoWait, option.Args)
	return delivery, err
}

// ConsumeOption ...
type ConsumeOption struct {
	Consumer  string
	AutoAck   bool
	Exclusive bool
	NoWait    bool
	Args      amqp.Table
}

// ConsumeOpt ...
type ConsumeOpt func(opt *ConsumeOption)

// ConsumerName sets consumers name.
// Consumer name is identified by a string that is unique and scoped for all
// consumers on this channel.  If you wish to eventually cancel the Consumer, use
// the same non-empty identifier in channel.Cancel.  An empty string will cause
// the library to generate a unique identity.  The Consumer identity will be
// included in every Delivery in the ConsumerTag field.
func ConsumerName(name string) ConsumeOpt {
	return func(option *ConsumeOption) {
		option.Consumer = name
	}
}

// AutoAck consume auto ack.
func AutoAck(autoAck bool) ConsumeOpt {
	return func(opt *ConsumeOption) {
		opt.AutoAck = autoAck
	}
}

// ConsumeExclusive consume auto ack.
func ConsumeExclusive(exclusive bool) ConsumeOpt {
	return func(opt *ConsumeOption) {
		opt.Exclusive = exclusive
	}
}

// ConsumeNoWait consume auto ack.
func ConsumeNoWait(noWait bool) ConsumeOpt {
	return func(opt *ConsumeOption) {
		opt.NoWait = noWait
	}
}
