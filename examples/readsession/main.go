package main

import (
	"bytes"
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx := context.Background()

	// Replace with one of your keys, e.g. "session_4KG4I..."
	var key string
	flag.StringVar(&key, "key", "", "Redis session key")
	flag.Parse()

	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		log.Fatal(err)
	}

	// If you stored complex/custom types in the session, register them first:
	// gob.Register(map[string]interface{}{}) // example if you used that type

	var values map[interface{}]interface{} // matches what the store saves
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&values); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("decoded session values: %#v\n", values)
}
