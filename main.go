package main

import (
	"context"
	grpcserver "controlserver/grpc_server"
	rt "controlserver/realtime"
	webserver "controlserver/web_server"
	"log"
	"os"
	"os/signal"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	connectionManager := rt.NewConnectionManager()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Print("Starting grpcserver")
		grpcserver.CreateServer(ctx, &connectionManager)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Print("Starting webserver")
		webserver.CreateServer(ctx, &connectionManager)
	}()

	wg.Wait()
}
