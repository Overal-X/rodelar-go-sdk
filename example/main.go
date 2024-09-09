package main

import (
	"context"
	"fmt"
	"log"

	"github.com/overal-x/rodelar-go-sdk"
)

func main() {
	client, err := rodelar.NewRodelarClient(
		rodelar.RodelarClientConfig{Url: "ws://localhost:3000/ws/"})
	if err != nil {
		log.Fatal(err)
	}

	err = client.Publish(rodelar.PublishArgs{
		Event:   "test-2",
		Message: "hello",
	})
	if err != nil {
		log.Fatal(err)
	}

	client.Publish(rodelar.PublishArgs{
		Event:   "test",
		Message: map[string]int{"x": 1, "y": 2, "z": 3},
	})

	err = client.Subscribe(rodelar.SubscribeArgs{
		Event: "test",
		Callback: func(m rodelar.Message) {
			fmt.Println("__", m.Event, m.Message)
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	client.Subscribe(rodelar.SubscribeArgs{
		Event: "test-1",
		Callback: func(m rodelar.Message) {
			fmt.Println("-> ", m.Event, m.Message)
		},
	})

	client.Subscribe(rodelar.SubscribeArgs{
		Event: "test-2",
		Callback: func(m rodelar.Message) {
			fmt.Println("-> ", m.Event, m.Message)
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	client.Listen(ctx)
	defer cancel()
}
