package sse

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

var clientConnectedCallbackCalls = 0
var clientDisconnectedCallbackCalls = 0

func clientConnected(c *Client) {
	clientConnectedCallbackCalls++
	log.Printf("client connected: %v", c)
}

func clientDisconnected(c *Client) {
	clientDisconnectedCallbackCalls++
	log.Printf("client disconnected: %v", c)
}

func TestNewServerNilOptions(t *testing.T) {
	srv := NewServer(nil)
	defer srv.Shutdown()

	if srv == nil || srv.options == nil || srv.options.Logger == nil {
		t.Fail()
	}
}

func TestNewServerNilLogger(t *testing.T) {
	srv := NewServer(&Options{
		Logger: nil,
	})

	defer srv.Shutdown()

	if srv == nil || srv.options == nil || srv.options.Logger == nil {
		t.Fail()
	}
}

func TestServer(t *testing.T) {
	channelCount := 2
	clientCount := 5
	messageCount := 0

	srv := NewServer(&Options{
		ClientConnectedFunc:    clientConnected,
		ClientDisconnectedFunc: clientDisconnected,
		Logger:                 log.New(os.Stdout, "go-sse: ", log.Ldate|log.Ltime|log.Lshortfile),
	})

	defer srv.Shutdown()

	// Create N channes
	for n := 0; n < channelCount; n++ {
		name := fmt.Sprintf("CH-%d", n+1)
		srv.addChannel(name)
		fmt.Printf("Channel %s registed\n", name)
	}

	wg := sync.WaitGroup{}
	m := sync.Mutex{}

	// Create N clients in all channes
	for n := 0; n < clientCount; n++ {
		for name, ch := range srv.channels {
			wg.Add(1)

			// Create new client
			c := newClient("", name)
			// Add client to current channel
			ch.addClient(c)

			id := fmt.Sprintf("C-%d", n+1)
			fmt.Printf("Client %s registed to channel %s\n", id, name)

			go func(id string) {
				// Wait for messages in the channel
				for msg := range c.send {
					m.Lock()
					messageCount++
					m.Unlock()
					fmt.Printf("Channel: %s - Client: %s - Message: %s\n", name, id, msg.data)
					wg.Done()
				}
			}(id)
		}
	}

	// Send hello message to all channels and all clients in it
	srv.SendMessage("", SimpleMessage("hello"))

	srv.close()

	wg.Wait()

	if messageCount != channelCount*clientCount {
		t.Errorf("Expected %d messages but got %d", channelCount*clientCount, messageCount)
	}
}

func TestServerAdditionalOptions(t *testing.T) {
	srv := NewServer(&Options{
		ClientConnectedFunc:    clientConnected,
		ClientDisconnectedFunc: clientDisconnected,
		Logger:                 log.New(os.Stdout, "go-sse: ", log.Ldate|log.Ltime|log.Lshortfile),
	})

	defer srv.Shutdown()

	http.Handle("/events/", srv)
	go func() {
		err := http.ListenAndServe("localhost:8081", nil)
		if err != nil {
			log.Printf("error running http server: %v", err)
		}
	}()

	client := http.Client{}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/events/asd", nil)
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("cannot get event url")
	}
	defer res.Body.Close()
	<-ctx.Done()
	time.Sleep(time.Millisecond * 300)
	if clientConnectedCallbackCalls != 1 {
		t.Errorf("clientConnectedCallbackCalls != 1")
	}
	if clientDisconnectedCallbackCalls != 1 {
		t.Errorf("clientDisconnectedCallbackCalls != 1")
	}
}
