package handlers

import (
	"fmt"

	"github.com/arizon-dread/sse/internal/config"
	"github.com/redis/go-redis"
	"github.com/redis/go-redis/v6"
)

type CacheMsgHandler struct {
	Name string
	Ch   chan string
	sub  *redis.PubSub
}

func (cmh CacheMsgHandler) Pub(msg string) error {
	rdb, err := conn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not publishing message, %v\n", err)
	}
	defer rdb.Close()
	err = rdb.Publish(cmh.Name, msg).Err()
	if err != nil {
		return err
	}
	return nil
}

func (cmh CacheMsgHandler) Sub(ch chan string) error {
	rdb, err := conn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not publishing message, %v\n", err)
	}
	cmh.sub = rdb.Subscribe(cmh.Name)

	for {
		msg, err := cmh.sub.ReceiveMessage()
		if err != nil {
			break
		}
		ch <- msg.Payload

	}
	return nil
}

func (cmh CacheMsgHandler) Exists() bool {
	if _, ok := <-cmh.Ch; !ok {
		return false
	}
	return true
}
func (cmh CacheMsgHandler) Unregister() {
	if cmh.sub != nil {
		cmh.sub.Unsubscribe(cmh.Name)
	}
}
func conn() (*redis.Client, error) {
	conf := config.Get()
	if conf.Cache.Url == "" {
		return nil, fmt.Errorf("no url to cache supplied in settings")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Cache.Url,
		Password: conf.Cache.Password,
	})

	return rdb, nil
}
