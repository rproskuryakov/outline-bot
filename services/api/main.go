package main

import (

    "net/http"
    "github.com/gorilla/mux"

    "github.com/rproskuryakov/outline-bot/services/api/internal/handlers"

)

func main() {

 // Create a new request multiplexer
 // Take incoming requests and dispatch them to the matching handlers
 router := mux.NewRouter()

 serverHandler := handlers.ServerHandler{}
 userHandler := handlers.UserHandler{}
 // Register the routes and handlers
 router.HandleFunc("/servers", serverHandler.CreateServer).Methods("POST")
 router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")

 // Run the server
 http.ListenAndServe(":8080", router)
}

