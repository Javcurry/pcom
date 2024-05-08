package iamqp_test

import (
	"hago-plat/pcom/iamqp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	assert.Equal(t, "amqp://hagopl:2JUNR9O4@rabbitmq-joyyfstest001-core001.duowan.com:8203/task_manager",
		iamqp.MakeURL("duowan", "duowan123", "183.36.110.69", "5672"))
}
