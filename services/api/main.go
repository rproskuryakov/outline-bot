package main

import (

    "net/http"

    "github.com/rproskuryakov/outline-bot/services/api/internal/handlers"

)

func main() {

 // Create a new request multiplexer
 // Take incoming requests and dispatch them to the matching handlers
 mux := http.NewServeMux()

 // Register the routes and handlers
 mux.Handle("/createServer", &CreateServerHandler{})
 mux.Handle("/createUser", &CreateUserHandler{})
 mux.Handle("/changeLimits")

 // Run the server
 http.ListenAndServe(":8080", mux)
}

