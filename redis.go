package main

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var db *redis.Client
var ctx = context.Background()

func connectRedis() {
	opts, err := redis.ParseURL(CFG_REDIS_URL)
	check(err)

	db = redis.NewClient(opts)
	pong, err := db.Ping(ctx).Result()
	check(err)

	if pong != "PONG" {
		log.Panicf("Could not communicate with Redis: %v", err)
	}
}

func closeRedis() {
	db.Close()
}
