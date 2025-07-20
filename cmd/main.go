package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Trip Plan Service is running!")
	})
	fmt.Println("Trip Plan Service listening on :8083")
	http.ListenAndServe(":8083", nil)
}
