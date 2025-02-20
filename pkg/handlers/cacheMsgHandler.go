package handlers

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/arizon-dread/sse/internal/config"
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

func (cmh CacheMsgHandler) Receive(ctx context.Context, ch chan string, cancel context.CancelFunc) error {
	rdb, err := helpers.GetCacheConn()
	if err != nil {
		return fmt.Errorf("unable to connect to redis, not receiving messages, %v", err)
	}
	cmd := rdb.Get(ctx, "receiver-"+cmh.Name)
	res, _ := cmd.Result()
	//If this is a new consumer that has messages waiting in redis streams, start by giving the consumer the last 2(configurable) messages in the order they were sent.
	if res == "+" || res == "" || res == "$" {
		go func() {
			cfg := config.Get()
			cacheCount := cfg.Cache.UnreadCacheCount
			if cacheCount == 0 {
				res = "$"
				rdb.Set(ctx, "receiver-"+cmh.Name, res, time.Hour*24)
				return
			}
			m := rdb.XRevRangeN(ctx, cmh.Name, "+", "-", int64(cacheCount))
			msgs, err := m.Result()
			if err != nil {
				return
			}
			slices.Reverse(msgs)
			for _, mess := range msgs {
				for _, vals := range mess.Values {
					switch v := vals.(type) {
					case string:
						ch <- v
						res = mess.ID
					}
				}
			}

			rdb.Set(ctx, "receiver-"+cmh.Name, res, time.Hour*24)
		}()
	}
	for {

		select {
		case <-ctx.Done():
			cancel()
			return nil
		default:
			var msg *redis.XStreamSliceCmd
			if res == "$" {
				msg = rdb.XRead(ctx, &redis.XReadArgs{Streams: []string{cmh.Name}, Block: 86400, Count: 1, ID: res})
			} else {
				msg = rdb.XRead(ctx, &redis.XReadArgs{Streams: []string{cmh.Name}, Block: -1, Count: 1, ID: res})
			}
			streams, err := msg.Result()
			if err != nil && err.Error() != "redis: nil" {
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
							rdb.Set(ctx, "receiver-"+cmh.Name, m.ID, time.Hour*24)
						default:
							return nil
						}
					}
				}
			}
		}
	}

}
func (cmh CacheMsgHandler) GetLastRead() *time.Time {
	rdb, err := helpers.GetCacheConn()
	if err != nil {
		return nil
	}
	res, _ := rdb.Get(context.Background(), "receiver-"+cmh.Name).Result()
	if res == "" {
		return nil
	}
	r := strings.Split(res, "-")
	i, err := strconv.ParseInt(r[0], 10, 64)
	if err != nil {
		panic(err)
	}
	tm := time.Unix(i/1000, 0)

	return &tm
}
func (cmh CacheMsgHandler) SetLastRead(d time.Time) {
	//this func is here to satisfy the interface, but read is saved when sending the message on the client.
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
