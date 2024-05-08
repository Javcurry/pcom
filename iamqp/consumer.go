package iamqp

import (
	"time"

	"github.com/streadway/amqp"
)

// Consumer ...
type Consumer struct {
	coroutinePool *Manager
	// conn          *amqp.Connection
	channel       *channel
	consumeOption *ConsumeOption
	queue         string
	handler       AMQPHandler
	done          chan bool
}

const (
	defaultConsumerTaskQueueSize   = 64
	defaultConsumerMaxCoroutineNum = 128
)

// NewConsumer ...
func NewConsumer(queue string, handler AMQPHandler, opt ...ConsumerOpt) *Consumer {
	option := &ConsumerOption{
		taskQueueSize:   defaultConsumerTaskQueueSize,
		maxCoroutineNum: defaultConsumerMaxCoroutineNum,
		consumeOption: &ConsumeOption{
			Consumer:  "",
			AutoAck:   true,
			Exclusive: false,
			NoWait:    false,
		},
	}
	for _, o := range opt {
		o(option)
	}
	consumer := &Consumer{
		queue:         queue,
		coroutinePool: newManager(option.taskQueueSize, option.maxCoroutineNum),
		consumeOption: option.consumeOption,
		handler:       handler,
	}
	if option.qos != nil {
		_ = consumer.Qos(option.qos.prefetchCount, option.qos.prefetchSize, option.qos.global)
	}
	return consumer
}

// ConsumerOption ...
type ConsumerOption struct {
	taskQueueSize   int
	maxCoroutineNum int
	consumeOption   *ConsumeOption
	qos             *qosParam
}

// ConsumerOpt ...
type ConsumerOpt func(opt *ConsumerOption)

// WithConsumeOption set Consumer's consume options.
func WithConsumeOption(consumeOpt *ConsumeOption) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.consumeOption = consumeOpt
	}
}

// ConsumerAutoAck set Consumer consume auto ack.
func ConsumerAutoAck(autoAck bool) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.consumeOption.AutoAck = autoAck
	}
}

// ConsumerExclusive set exclusive in consumer
func ConsumerExclusive(exclusive bool) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.consumeOption.Exclusive = exclusive
	}
}

// ConsumerNoWait set auto ack in consumer.
func ConsumerNoWait(noWait bool) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.consumeOption.NoWait = noWait
	}
}

// SetConsumerTaskQueueSize ...
func SetConsumerTaskQueueSize(size int) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.taskQueueSize = size
	}
}

// SetMaxCoroutineNum ...
func SetMaxCoroutineNum(num int) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.maxCoroutineNum = num
	}
}

// SetQos ...
func SetQos(prefetchCount, prefetchSize int, global bool) ConsumerOpt {
	return func(opt *ConsumerOption) {
		opt.qos = &qosParam{
			prefetchCount: prefetchCount,
			prefetchSize:  prefetchSize,
			global:        global,
		}
	}
}

type qosParam struct {
	prefetchCount, prefetchSize int
	global                      bool
}

// Qos set Qos of consumer. Since channel in iamqp has only one consumer, global
// parameter is useless when we use RabbitMQ. (see https://www.rabbitmq.com/consumer-prefetch.html)
func (c *Consumer) Qos(prefetchCount, prefetchSize int, global bool) error {
	return c.channel.ch.Qos(prefetchCount, prefetchSize, global)
}

func (c *Consumer) serve(ch *channel) {
	c.coroutinePool.run()

	c.done = make(chan bool, 1)
	c.channel = ch
	needReInit := false
	first := true
	var delivery <-chan amqp.Delivery
	for {
		if needReInit || first {
			first = false
			needReInit = false
			err := ch.initConsumeChannel()
			if err != nil {
				Log("initConsumeChannel %v", err)
			}
			delivery, err = ch.consume(c.queue, c.consumeOption)
			if err != nil {
				Log("consume err: %v", err)
				select {
				case <-c.done:
					return
				case <-time.After(reInitDelay):
					continue
				}
			}
		}
		select {
		case <-c.channel.notifyClose:
			needReInit = true
		case d, ok := <-delivery:
			if !ok {
				needReInit = true
				continue
			}
			task := &consumeTask{
				msg:     message(d),
				handler: c.handler,
			}
			c.coroutinePool.Dispatch(task)
		case <-c.done:
			return
		}
	}
}

// Done stops consumer
func (c *Consumer) Done() {
	c.done <- true
}

type consumeTask struct {
	msg     *Message
	handler AMQPHandler
}

// Launch ...
func (t *consumeTask) Launch() {
	t.handler(t.msg)
}

// AMQPHandler ...
type AMQPHandler func(msg *Message)
