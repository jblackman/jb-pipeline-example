package main

// This is a simple API

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	r := handler()
	log.Fatal(http.ListenAndServe(":"+getEnv("PORT", "8080"), r))
}

func handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/greeting", greetingHandler)
	return r
}

func getEnv(name, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}
	return value
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	greeting := getEnv("GREETING", "Hello world")
	fmt.Fprintf(w, greeting)
}

type GreetingJSON struct {
	Greeting string `json:"greeting"`
}

func greetingHandler(w http.ResponseWriter, r *http.Request) {
	greeting := GreetingJSON{Greeting: getEnv("GREETING", "not set")}
	bytes, _ := json.Marshal(&greeting)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}
