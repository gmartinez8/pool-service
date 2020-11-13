package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gmartinez8/server"
)

//This is an example of server that is listening to a callback of the pooltask service
//Listen on port 8080
//Each time the callback is called will log on console: {ID, Success}
//{ID:47ad1b18d923c4315058a4798ac41507 Success:true}
func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	//for cancelation propagation
	//Background because its the root - there is nothing previous
	ctx, cancel := context.WithCancel(context.Background())
	//Its important to call cancel() becauset it releases resources associated with the context
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)
	s := server.NewServer(":8080")
	s.Handle("/callback", "POST", HandleCallback)

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

//CallbackRequest for expected response after creating a task
type CallbackRequest struct {
	ID      string `json:"taskID"`
	Success bool   `json:"success"`
}

//HandleCallback handles all finished tasks
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cr CallbackRequest
	err := decoder.Decode(&cr)
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		e, _ := json.Marshal(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(e)
		return
	}
	response, err := json.Marshal(cr)
	if err != nil {
		e, _ := json.Marshal(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(e)
		return
	}
	log.Printf("%+v\n", cr)
	w.WriteHeader(http.StatusOK)
	w.Write(response)
}
