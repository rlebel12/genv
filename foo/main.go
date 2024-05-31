package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	lambda.Start(mux)
}
