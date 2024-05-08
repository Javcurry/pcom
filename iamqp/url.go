package iamqp

import (
	"net"
	"net/url"
)

func newURL(username, passwd, host, port string) *url.URL {
	return &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(username, passwd),
		Host:   net.JoinHostPort(host, port),
		Path:   "/",
	}
}

// MakeURL 生成amqp连接url
func MakeURL(username, passwd, host, port string) string {
	return newURL(username, passwd, host, port).String()
}
