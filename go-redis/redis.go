package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"sync"
	"time"
)

type collector struct {
	sync.Mutex
	items []time.Duration
}

func NewCollector() collector {
	return collector{
		items: make([]time.Duration, 0, 8),
	}
}
func (c *collector) Append(d time.Duration) {
	c.Lock()
	defer c.Unlock()

	c.items = append(c.items, d)
}
func (c *collector) Print() {
	if len(c.items) == 0 {
		return
	}

	var sum time.Duration
	var max time.Duration
	var min time.Duration = c.items[0]

	for _, item := range c.items {
		sum += item

		if max < item {
			max = item
		}

		if min > item {
			min = item
		}
	}

	avg := sum / time.Duration(len(c.items))

	log.Println("TOTAL:", len(c.items))
	log.Println("MAX  :", max)
	log.Println("MIN  :", min)
	log.Println("AVG  :", avg)
}

const (
	semaphoreCount = 8
)

var (
	collection collector
	semaphore  chan struct{}
)

func init() {
	collection = NewCollector()
	semaphore = make(chan struct{}, semaphoreCount)
}

func main() {
	log.Println("START")
	defer log.Println("END")

	defer collection.Print()

	const dataSize = 600000
	const loopCount = 100
	const key = "test-key"

	list := make([]byte, 0, dataSize)
	for i := 0; i < dataSize; i++ {
		list = append(list, '0')
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: ":6379",
		DB:   0,
	})

	ctx := context.Background()

	err := rdb.Set(ctx, key, string(list), 0).Err()
	if err != nil {
		log.Fatal(err)
	}

	massiveGet(ctx, rdb, key, loopCount)
}

func massiveGet(ctx context.Context, rdb *redis.Client, key string, loopCount int) {
	var wg sync.WaitGroup
	for i := 0; i < loopCount; i++ {
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer func() {
				wg.Done()
				<-semaphore
			}()
			t := time.Now()
			_, err := rdb.Get(ctx, key).Result()
			if err != nil {
				log.Fatal(err)
			}
			collection.Append(time.Now().Sub(t))
		}()
	}
	wg.Wait()
}
