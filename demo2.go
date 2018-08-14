package main

import (
	"fmt"
	"github.com/jpillora/overseer"
	"net/http"
	"os"
	"time"
)

func prog(state overseer.State) {
	fmt.Printf("app#%s (%s) listening...\n", os.Getppid(), state.ID)
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		duration, err := time.ParseDuration(r.FormValue("duration"))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		time.Sleep(duration)
		w.Write([]byte("Hello World"))
		fmt.Fprintf(w, "app#%s (%s) says hello\n", os.Getppid(), state.ID)
	}))
	http.Serve(state.Listener, nil)
	fmt.Printf("app#%s (%s) exiting...\n", os.Getppid(), state.ID)
}

func main() {
	overseer.Run(overseer.Config{
		Program:   prog,
		Addresses: []string{":5005", ":5006"},
		Debug:     false, //display log of overseer actions
	})
}
