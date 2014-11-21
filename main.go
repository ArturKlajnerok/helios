package main

import (
	"net/http"
)

func tokenHandler(w http.ResponseWriter, r *http.Request) {
}

func main() {

	http.HandleFunc("/token", tokenHandler)
	http.ListenAndServe(":8080", nil)
}
