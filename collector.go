package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

// Logging handler
type LogWriter struct {
}

func (writer LogWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + " [DEBUG] " + string(bytes))
}

// Lambda handler struct
type MyEvent struct {
	Name string `json:"name"`
}

// Lambda handler func
func handleRequest(ctx context.Context, name MyEvent) (string, error) {
	queryApiChannels()
	return fmt.Sprintf("Hello %s!", name.Name), nil
}

func main() {
	// Set logging parameters
	log.SetFlags(0)
	log.SetOutput(new(LogWriter))
	log.Println("status=collection_process_started message=started")

	// Auto-detect environment for osx, launch non-lambda, for testing
	log.Printf("status=os_detected message=%s\n", runtime.GOOS)
	if runtime.GOOS == "darwin" {
		queryApiChannels()
		// Auto-detect environment for linux (aws lambda), launch lambda
	} else if runtime.GOOS == "linux" {
		lambda.Start(handleRequest)
	}
}
