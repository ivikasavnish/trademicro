package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ably/ably-go/ably"
	"github.com/joho/godotenv"
)

type ProcessCommand struct {
	Action string   `json:"action"`
	Script string   `json:"script"`
	Args   []string `json:"args"`
}

type ProcessStatus struct {
	Status string   `json:"status"`
	Script string   `json:"script"`
	Args   []string `json:"args"`
	Pid    int      `json:"pid"`
	Error  string   `json:"error"`
}

func main() {
	// Load .env file for ABLY_API_KEY and ABLY_CHANNEL
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}
	apiKey := os.Getenv("ABLY_API_KEY")
	channelName := os.Getenv("ABLY_CHANNEL")
	if apiKey == "" || channelName == "" {
		log.Fatal("Missing ABLY_API_KEY or ABLY_CHANNEL env var")
	}

	// Use Ably Realtime for both publish and subscribe
	client, err := ably.NewRealtime(ably.WithKey(apiKey))
	if err != nil {
		log.Fatalf("Ably Realtime connection error: %v", err)
	}
	ch := client.Channels.Get(channelName)

	// Send a start command
	cmd := ProcessCommand{
		Action: "start",
		Script: "trade_log.py",
		Args:   []string{"COALINDIA", "5", "--diff", ".1"},
	}
	cmdJson, _ := json.Marshal(cmd)
	if err := ch.Publish(context.Background(), "command", string(cmdJson)); err != nil {
		log.Fatalf("Failed to publish: %v", err)
	}
	fmt.Println("Published start command")

	// Subscribe for status updates using a callback
	statusCh := make(chan ProcessStatus, 10)
	_, err = ch.Subscribe(context.Background(), "status", func(msg *ably.Message) {
		var status ProcessStatus
		if err := json.Unmarshal([]byte(msg.Data.(string)), &status); err == nil {
			statusCh <- status
		}
	})
	if err != nil {
		log.Fatalf("Subscribe error: %v", err)
	}
	fmt.Println("Waiting for status updates... (Press Ctrl+C to exit)")
	// Listen indefinitely for status updates
	for status := range statusCh {
		fmt.Printf("Status: %s, Script: %s, Pid: %d, Error: %s\n", status.Status, status.Script, status.Pid, status.Error)
	}
}
