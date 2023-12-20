### Example

#### For rocketmq

```go
package main

// Package main implements a simple producer to send message.
func main() {
	addrs := []string{"127.0.0.1:9876"}
	
	p, _ := rocketmq.NewProducer(
		producer.WithNsResolver(NewCacheResolver(addrs,
			WithCacheFirst(true),
			WithIpStorageFirst(true),
			WithDnsTimeout(1*time.Second),
		)),
		producer.WithRetry(2),
	)
	err := p.Start()
	if err != nil {
		fmt.Printf("start producer error: %s", err.Error())
		os.Exit(1)
	}
	topic := "TopicTest"

	for i := 0; i < 10; i++ {
		msg := &primitive.Message{
			Topic: topic,
			Body:  []byte("Hello RocketMQ Go Client! " + strconv.Itoa(i)),
		}
		res, err := p.SendSync(context.Background(), msg)

		if err != nil {
			fmt.Printf("send message error: %s\n", err)
		} else {
			fmt.Printf("send message success: result=%s\n", res.String())
		}
	}
	err = p.Shutdown()
	if err != nil {
		fmt.Printf("shutdown producer error: %s", err.Error())
	}
}


```


#### For redis
```go

package main

func ExampleNewClient() {
	opt := &redis.Options{
		Addr:     "localhost:6379", // use default Addr
	}

	opt.Dialer = NewDialerWithCache(opt,
		WithCacheFirst(true),
		WithIpStorageFirst(true),
		WithDnsTimeout(2*time.Second),
	)
	rdb := redis.NewClient(opt)

	pong, err := rdb.Ping(ctx).Result()
	fmt.Println(pong, err)
	// Output: PONG <nil>
}


```

