package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arizon-dread/sse/internal/helpers"
	"github.com/redis/go-redis/v9"
)

type CacheMsgHandler struct {
	Name string
	Ch   chan string
}

func (cmh CacheMsgHandler) Send(msg string) error {
	rdb, err := helpers.GetCacheConn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not publishing message, %v", err)
	}
	defer rdb.Close()
	ctx := context.Background()
	messageData := make(map[string]string)
	messageData["data"] = msg
	cmd := rdb.XAdd(ctx, &redis.XAddArgs{Stream: cmh.Name, ID: "*", Approx: true, MaxLen: 100, Values: messageData})
	res, err := cmd.Result()
	if err != nil {
		log.Printf("Got err when adding to stream, %v\n", err)
		return err
	}
	log.Printf("Added %v to stream %v and got id %v", msg, cmh.Name, res)
	return nil
}

func (cmh CacheMsgHandler) Receive(ctx context.Context, ch chan string) error {
	rdb, err := helpers.GetCacheConn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not receiving messages, %v", err)
	}
	cmd := rdb.Get(ctx, "receiver-"+cmh.Name)
	res, _ := cmd.Result()
	if res == "" {
		res = "$"
	}
	for {

		select {
		case <-ctx.Done():
			return nil
		default:

			msg := rdb.XRead(ctx, &redis.XReadArgs{Streams: []string{cmh.Name}, Count: 2, ID: res})
			streams, err := msg.Result()
			if err != nil {
				ctx.Done()
				return fmt.Errorf("reading failed, %v", err)
			}
			for _, v := range streams {
				for _, m := range v.Messages {
					for _, val := range m.Values {
						switch v := val.(type) {
						case string:
							ch <- v
							res = m.ID
							rdb.Set(ctx, "receiver-"+cmh.Name, m.ID, time.Hour*25)
						default:
							return nil
						}
					}
				}
			}
		}
	}

}

func (cmh CacheMsgHandler) Exists() bool {
	rdb, err := helpers.GetCacheConn()
	if err != nil {
		return false
	}
	info := rdb.XInfoStream(context.Background(), cmh.Name)
	return info.Val().Length > 0
}
func (cmh CacheMsgHandler) Unregister() {
	log.Printf("Unregistering client %v\n", cmh.Name)
	close(cmh.Ch)
}

func (cmh CacheMsgHandler) GetName() string {
	return cmh.Name
}

func (cmh CacheMsgHandler) GetChannel() chan string {
	return cmh.Ch
}
