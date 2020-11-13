package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gmartinez8/pooltask"
	"github.com/gmartinez8/server"
)

//This is an example of how to use pooltask mod
//we can SetMaxWorkers number
//we can set CallbackURL
func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//Default workers is 10 we can set the max number to 6 to SetMaxWorkers
	pooltask.SetMaxWorkers(6)
	//pooltask.CallbackURL = "http://localhost:8081/callback"
	//for cancelation propagation
	//Background because its the root - there is nothing previous
	ctx, cancel := context.WithCancel(context.Background())
	//Its important to call cancel() becauset it releases resources associated with the context
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	s := server.NewServer(":4000")
	s.Handle("/", "GET", pooltask.HandleHome)
	s.Handle("/task", "GET", pooltask.HandleListTasks)
	s.Handle("/task", "POST", pooltask.HandleCreateTask)

	go func(ctx context.Context) {
		defer wg.Done()
		if err := s.Run(ctx); err != nil {
			log.Fatalf("service start failed: %v", err)
		}
	}(ctx)

	<-ch
	log.Println("server service is shutting down...")
	cancel()
	wg.Wait()
	log.Println("shut down finished")
}
