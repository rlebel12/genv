package main

import (
	"log/slog"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	slog.Info("Server started on port 8080")
	http.ListenAndServe(":8080", mux)
}
