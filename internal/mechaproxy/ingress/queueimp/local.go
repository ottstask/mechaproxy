package queueimp

import (
	"net/url"
)

type localImp struct {
	ch chan *Message
}

func newLocalImp(u *url.URL) (*localImp, error) {
	return &localImp{ch: make(chan *Message, 1000)}, nil
}

func (l *localImp) Push(m *Message) {
	l.ch <- m
}

func (l *localImp) Queue() <-chan *Message {
	return l.ch
}
