package queueimp

import (
	"fmt"
	"net/url"
)

type MessageQueue interface {
	Push(*Message)
	Queue() <-chan *Message
}

type Message struct {
	CallbackAddr string
	Payload      []byte
}

func NewQueueImp(source string) (MessageQueue, error) {
	u, err := url.Parse(source)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "redis":
		return newRedisImp(u)
	case "local":
		return newLocalImp(u)
	default:
		return nil, fmt.Errorf("unsupport queue schema %s", u.Scheme)
	}
}
