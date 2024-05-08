package iamqp_test

import (
	"fmt"
	"hago-plat/pcom/iamqp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPubSub(t *testing.T) {
	queue := "test_queue"
	queue2 := "test_queue2"
	exchange := "test_iamqp"
	routeKey := "rk_iamqp"
	pubSes, err := iamqp.NewSession(iamqp.MakeURL("duowan", "duowan123", "183.36.110.69", "5672"), nil)
	err = pubSes.ExchangeDeclare(exchange, iamqp.ExchangeTypeDirect)
	assert.NoError(t, err)

	subSes, err := iamqp.NewSession(iamqp.MakeURL("duowan", "duowan123", "183.36.110.69", "5672"), nil)
	subSes.Consume(iamqp.NewConsumer(queue, func(msg *iamqp.Message) {
		fmt.Println("msg:", string(msg.Body))
	}))
	err = subSes.BindAndConsume(queue2, exchange, routeKey, func(msg *iamqp.Message) {
		// fmt.Println("msg:", string(msg.Body))
	})
	t.Log(err)
	go func() {
		for {
			select {
			case <-time.After(10 * time.Millisecond):
				pubSes.Publish(exchange, routeKey, iamqp.Message{Body: []byte("hello!!")})
			case <-time.After(30 * time.Second):
				return
			}
		}
	}()
	t5 := time.After(5 * time.Second)
	//t15 := time.After(15 * time.Second)
	runningTime := time.After(5 * time.Minute)
	for {
		select {
		case <-t5:
			_, err = subSes.QueueDeclare(queue)
			assert.NoError(t, err)
			err = subSes.QueueBind(queue, routeKey, exchange)
			assert.NoError(t, err)
			subSes.Consume(iamqp.NewConsumer(queue, func(msg *iamqp.Message) {
				// fmt.Println("msg2:", string(msg.Body))

			}))
		//case <-t15:
		//	err = subSes.QueueUnbind(queue, routeKey, exchange, nil)
		//	assert.NoError(t, err)
		//	_, err = subSes.QueueDelete(queue)
		//	assert.NoError(t, err)
		case <-runningTime:
			return
		}
	}
}
