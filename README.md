# Rate Limiter Implementation

This program is a simple implementation using Golang and Redis.
This program utilize `&sync.Mutex{}` and `map` as a rate limiter storage.

## Run The Program

```shell
$ go run main.go
2022/06/09 15:17:55 Connected to redis at: localhost:6379
2022/06/09 15:17:55 Listening at         : 8000
2022/06/09 15:17:55 Window time          : 1 s
2022/06/09 15:17:55 Max Request          : 5
```

## Available Configuration

```shell
$ go run main.go -h
Usage of main:
  -max-request int
    	The rate limit max requests within the window (number of requests) (default 5)
  -redis-address string
    	The redis address where we want to get the value from (default "localhost:6379")
  -server-port int
    	The server port where you want to listen and serve this program (default 8000)
  -window-size int
    	The rate limit window size (in seconds) (default 1)
```

## Try to simulate Request

```shell
$ ./simulate.sh asd # will show the curl result
```