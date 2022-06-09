package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-redis/redis/v8"
	"log"
	"net/http"
	"sync"
	"time"
)

/*
You are working on a service which expects to handle a significant amount of traffic.
To ensure that the service stays up even when there are spikes in traffic,
you should devise a rate limiting function that returns true or false
depending on the number of requests that are coming in within a time window.

The function should consider 2 constants when making a decision:
The rate limit window size (in seconds)
The rate limit max requests within the window (number of requests)
*/

const (
	defaultRedisAddress       string = "localhost:6379"
	defaultMaxRequest         int    = 5
	defaultServerPort         int    = 8000
	defaultWindowSizeInSecond int    = 1
)

var (
	redisAddress        = flag.String("redis-address", defaultRedisAddress, "The redis address where we want to get the value from")
	windowSizeInSecond  = flag.Int("window-size", defaultWindowSizeInSecond, "The rate limit window size (in seconds)")
	maxRequestPerWindow = flag.Int("max-request", defaultMaxRequest, "The rate limit max requests within the window (number of requests)")
	serverPort          = flag.Int("server-port", defaultServerPort, "The server port where you want to listen and serve this program")
	mtx                 = &sync.Mutex{}
	counterMap          = make(map[string]int)
)

func rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mtx.Lock()
		defer mtx.Unlock()
		c := counterMap[r.URL.Query().Get("id")]
		c++
		counterMap[r.URL.Query().Get("id")] = c

		if c > *maxRequestPerWindow {
			limitedHandler(w, r)
			return
		}

		next(w, r)
	}
}

func rootHandler(rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		get := rdb.Get(r.Context(), r.URL.Query().Get("id"))
		if err := get.Err(); err != nil {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "NOT FOUND")
			return
		}

		val := get.String()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, val)
	}
}

func limitedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, r.URL.Query().Get("id"))
}

func getRedisClient(ctx context.Context, address string) (rdb *redis.Client, closer func(), err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: address,
	})

	if err = rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, err
	}

	return rdb, func() {
		if err := rdb.Close(); err != nil {
			panic(err)
		}
	}, nil
}

func seedRedis(ctx context.Context, rdb *redis.Client) error {
	for _, key := range []string{"asd", "asdf"} {
		set := rdb.Set(ctx, key, 1, 0)
		if err := set.Err(); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	rdb, closeRdb, err := getRedisClient(context.Background(), *redisAddress)
	if err != nil {
		panic(err)
	}
	defer closeRdb()

	if err := seedRedis(context.Background(), rdb); err != nil {
		panic("failed to seed redis:" + err.Error())
	}

	go func() {
		for range time.Tick(time.Duration(*windowSizeInSecond) * time.Second) {
			mtx.Lock()
			for k := range counterMap {
				counterMap[k] = 0
			}
			mtx.Unlock()
		}
	}()

	log.Println("Connected to redis at:", *redisAddress)
	log.Println("Listening at         :", *serverPort)
	log.Println("Window time          :", *windowSizeInSecond, "s")
	log.Println("Max Request          :", *maxRequestPerWindow)
	http.ListenAndServe(fmt.Sprintf(":%d", *serverPort), rateLimitMiddleware(rootHandler(rdb)))
}
