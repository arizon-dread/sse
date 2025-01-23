package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arizon-dread/sse/internal/config"
	"github.com/redis/go-redis/v9"
)

type CacheMsgHandler struct {
	Name string
	Ch   chan string
}

func (cmh CacheMsgHandler) Send(msg string) error {
	rdb, err := conn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not publishing message, %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()
	cmd := rdb.XAdd(ctx, &redis.XAddArgs{Stream: cmh.Name, MaxLen: 100, Values: msg})
	log.Printf("Added %v to stream %v and got id %v", msg, cmh.Name, cmd)
	return nil
}

func (cmh CacheMsgHandler) Receive(ch chan string) error {
	rdb, err := conn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not publishing message, %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	for {

		msg := rdb.XRead(ctx, &redis.XReadArgs{Streams: []string{cmh.Name}, Count: 2, Block: time.Duration(time.Second * 20)})
		streams, err := msg.Result()
		if err != nil {
			cancel()
			return fmt.Errorf("reading failed, %v", err)
		}
		for _, v := range streams {
			for _, m := range v.Messages {
				for _, val := range m.Values {
					switch v := val.(type) {

					case string:
						ch <- v
					default:
						break
					}
				}
			}
		}

	}
}

func (cmh CacheMsgHandler) Exists() bool {
	rdb, err := conn()
	if err != nil {
		return false
	}
	info := rdb.XInfoStream(context.Background(), cmh.Name)
	if info.Val().Length > 0 {
		return true
	}
	return false
}
func (cmh CacheMsgHandler) Unregister() {
	close(cmh.Ch)
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
