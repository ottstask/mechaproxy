package queueimp

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisImp struct {
	db  *redis.Client
	key string
}

func newRedisImp(u *url.URL) (*redisImp, error) {
	password, _ := u.User.Password()
	paths := strings.Split(u.Path, "/")
	if len(paths) != 3 {
		return nil, fmt.Errorf("bad path %s (format should be: /db/key)", u.Path)
	}
	db, _ := strconv.ParseInt(paths[1], 10, 32)
	key := paths[2]
	rdb := redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Password: password,
		DB:       int(db),
	})

	return &redisImp{db: rdb, key: key}, nil
}

func (r *redisImp) Push(m *Message) {
	r.db.LPush(context.Background(), r.key, m)
}

func (r *redisImp) Queue() <-chan *Message {
	ch := make(chan *Message, 100)
	go func() {
		for {
			m := &Message{}
			if err := r.db.LPop(context.Background(), r.key).Scan(m); err != nil {
				log.Println("pop scan error", err)
				time.Sleep(time.Second)
			} else {
				ch <- m
			}
		}
	}()
	return ch
}
