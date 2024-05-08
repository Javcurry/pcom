package iamqp

import (
	"crypto/tls"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Session ...
type Session struct {
	conn            *amqp.Connection
	publishChannel  chan *channel
	notifyConnClose chan *amqp.Error
	consumers       map[*Consumer]struct{}
	// addrURL         string
	// amqps      *tls.Config
	maxChannel int // todo动态调整channel池中channel个数
	isReady    bool
	reliable   bool
	ready      chan bool
	done       chan bool
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When setting up the channel after a channel exception
	reInitDelay = 2 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 5 * time.Second
)

const (
	// defaultMaxChan
	defaultMaxChan = 32
)

// NewSession return MQSession, MQSession doesn't provide connection pool.
// It contains publish channel pool.
func NewSession(amqpURL string, amqps *tls.Config) (*Session, error) {
	ses := &Session{
		done:       make(chan bool),
		reliable:   false,
		maxChannel: defaultMaxChan,
		ready:      make(chan bool, 1),
	}
	go ses.handleReconnect(amqpURL, amqps)
	isReady := <-ses.ready
	if !isReady {
		ses.done <- true
		return nil, errors.New("connect fail")
	}
	return ses, nil
}

// Confirm set session into confirm mode
func (s *Session) Confirm() {
	s.reliable = true
}

// Consume used to declare consumers
func (s *Session) Consume(cons *Consumer) error {
	if s.consumers == nil {
		s.consumers = make(map[*Consumer]struct{})
	}
	s.consumers[cons] = struct{}{}
	ch, err := newConsumerChannel(s)
	if err != nil {
		Log("iamqp: Consumer fail. err: %v", err)
		return err
	}
	go cons.serve(ch)
	return nil
}

// BindAndConsume auto declare queue and bind queue to exchange with routeKey before consume using default
// QueueDeclOption and default QBindOption.
func (s *Session) BindAndConsume(queue, exchange, routeKey string, handler AMQPHandler, opt ...ConsumerOpt) error {
	if s.consumers == nil {
		s.consumers = make(map[*Consumer]struct{})
	}
	_, err := s.QueueDeclare(queue)
	if err != nil {
		return err
	}
	err = s.QueueBind(queue, routeKey, exchange)
	if err != nil {
		return err
	}
	cons := NewConsumer(queue, handler, opt...)
	s.consumers[cons] = struct{}{}
	ch, err := newConsumerChannel(s)
	if err != nil {
		Log("iamqp: BindAndConsume fail. err: %v", err)
		return err
	}
	go cons.serve(ch)
	return nil
}

// Get get one msg in queue
func (s *Session) Get(queue string, autoAck bool) (*Message, error) {
	ch, err := s.conn.Channel()
	if err != nil {
		return nil, err
	}
	defer func() { _ = ch.Close() }()
	delivery, ok, err := ch.Get(queue, autoAck)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	msg := message(delivery)
	return msg, nil
}

// handleReconnect will wait for a connection error on
// notifyConnClose, and then continuously attempt to reconnect.
func (s *Session) handleReconnect(addr string, amqps *tls.Config) {
	for {
		s.isReady = false
		Log("iamqp: Attempting to connect")

		err := s.connect(addr, amqps)
		if err != nil {
			Log("iamqp: Failed to connect (url: %v). Retrying...", addr)
			select {
			case s.ready <- false:
			default:
			}

			select {
			case <-s.done:
				return
			case <-time.After(reconnectDelay):
			}
			continue
		}

		if done := s.handleReInit(); done {
			break
		}
	}
}

// connect will create a new AMQP connection
func (s *Session) connect(addr string, amqps *tls.Config) error {
	conn, err := amqp.DialTLS(addr, amqps)
	if err != nil {
		return err
	}
	s.changeConnection(conn)
	Log("iamqp: Connected! (not error)")
	return nil
}

// changeConnection takes a new connection to the queue,
// and updates the close listener to reflect this.
func (s *Session) changeConnection(connection *amqp.Connection) {
	s.conn = connection
	s.notifyConnClose = s.conn.NotifyClose(make(chan *amqp.Error, 1))
}

// handleReconnect will wait for a connect error
// and then continuously attempt to re-initialize
func (s *Session) handleReInit() bool {
	s.isReady = false
	s.publishChannel = make(chan *channel, s.maxChannel)
	for i := 0; i < s.maxChannel; i++ {
		newPubCh, err := newChannel(s)
		if err != nil {
			Log("iamqp: return handleReInit...")
			return false
		}
		s.publishChannel <- newPubCh
	}
	for cons := range s.consumers {
		ch, err := newConsumerChannel(s)
		if err != nil {
			Log("iamqp: new Consumer chan fail. err: %v", err)
			return false
		}
		go cons.serve(ch)
	}
	s.waitInFirstInit()
	Log("iamqp: connect started...")
	return s.waitClose()
}

func (s *Session) waitInFirstInit() {
	select {
	case s.ready <- true:
	default:
	}
}

func (s *Session) waitClose() bool {
	select {
	case <-s.done:
		s.closePublishChannel()
		s.closeConsumerChannel()
		Log("iamqp: Connection done...")
		return true
	case err := <-s.notifyConnClose:
		s.closePublishChannel()
		s.closeConsumerChannel()
		Log("iamqp: Connection closed. Reconnecting... err:%v", err)
		return false
	}
}

func (s *Session) closePublishChannel() {
	for {
		select {
		case c := <-s.publishChannel:
			c.Done()
		case <-time.After(5 * time.Second):
			close(s.publishChannel)
			return
		}
	}
}

func (s *Session) closeConsumerChannel() {
	for cons := range s.consumers {
		cons.Done()
	}
}

func (s *Session) publisher() (*channel, error) {
	pubCh := <-s.publishChannel
	if pubCh == nil || !pubCh.available.Load().(bool) {
		newPubCh, err := newChannel(s)
		if err != nil {
			Log("iamqp: get new channel fail. err:%v", err)
			return nil, err
		}
		return newPubCh, nil
	}
	return pubCh, nil
}
