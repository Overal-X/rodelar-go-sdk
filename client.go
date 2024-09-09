package rodelar

import (
	"context"
	"encoding/json"
	"log"

	"golang.org/x/net/websocket"
)

type IRodelarClient interface {
	Publish(PublishArgs) error
	Subscribe(SubscribeArgs) error
	Listen(ctx context.Context)
}

type RodelarClient struct {
	ws       *websocket.Conn
	handlers map[string]func(Message)
}

type RodelarClientConfig struct {
	Url      string
	ApiKeyId *string
	ApiKey   *string
}

type ActionEnum string

const (
	SUBSCRIBE ActionEnum = "SUBSCRIBE"
	PUBLISH   ActionEnum = "PUBLISH"
)

type Message struct {
	Action  ActionEnum  `json:"action"`
	Event   string      `json:"event"`
	Message interface{} `json:"message,omitempty"`
}

type PublishArgs struct {
	Event   string
	Message interface{}
}

func (r *RodelarClient) Publish(args PublishArgs) error {
	dataByte, err := json.Marshal(&Message{
		Action:  PUBLISH,
		Event:   args.Event,
		Message: args.Message,
	})
	if err != nil {
		return err
	}

	_, err = r.ws.Write(dataByte)
	return err
}

type SubscribeArgs struct {
	Event    string
	Callback func(Message)
}

func (r *RodelarClient) Subscribe(args SubscribeArgs) error {
	dataByte, err := json.Marshal(&Message{
		Action: SUBSCRIBE,
		Event:  args.Event,
	})
	if err != nil {
		return err
	}

	_, err = r.ws.Write(dataByte)
	if err != nil {
		return err
	}

	r.handlers[args.Event] = args.Callback
	return nil
}

// Listen reads messages and invokes the appropriate callback. It respects context cancellation.
func (r *RodelarClient) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done(): // If the context is cancelled, stop the loop
			log.Println("Stopping Listen loop.")
			r.ws.Close()
			return
		default:
			// TODO: server sudden closure
			// BUG: somehow, when ctx is cancel, the subscribers are still active in background
			msg := make([]byte, 1024) // Adjust buffer size based on expected message size
			n, err := r.ws.Read(msg)
			if err != nil {
				log.Println("WebSocket Read Error:", err)
				continue
			}

			msgStruct := Message{}
			err = json.Unmarshal(msg[:n], &msgStruct)
			if err != nil {
				log.Println("JSON Unmarshal Error:", err)
				continue
			}

			// Check if a handler is registered for the received event
			if handler, ok := r.handlers[msgStruct.Event]; ok {
				handler(msgStruct)
			} else {
				log.Printf("No handler for event: %s\n", msgStruct.Event)
			}
		}
	}
}

func NewRodelarClient(config RodelarClientConfig) (IRodelarClient, error) {
	ws, err := websocket.Dial(config.Url, "", "http://localhost/")
	if err != nil {
		return nil, err
	}

	return &RodelarClient{
		ws:       ws,
		handlers: make(map[string]func(Message)),
	}, nil
}
