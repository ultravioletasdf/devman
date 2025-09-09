package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is online")
		fmt.Println("Received request")
	})
	time.AfterFunc(time.Second*10, func() {
		panic("10 seconds")
	})
	panic(http.ListenAndServe("localhost:3000", nil))
}
