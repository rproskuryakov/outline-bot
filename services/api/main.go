package main

import (

    "os"
    "net/http"
    "database/sql"

    "github.com/gorilla/mux"
    "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"

    "github.com/rproskuryakov/outline-bot/services/api/internal/handlers"
    "github.com/rproskuryakov/outline-bot/services/api/internal/repositories"

)

func main() {
    var postgresDsn string = os.Getenv("POSTGRES_DSN")

    sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(postgresDsn)))
    db := bun.NewDB(sqldb, pgdialect.New())
//     err := db.ResetModel(ctx, (*model.User)(nil))
//     log.Printf("Table Users created")
//     if err != nil {
//         panic(err)
//     }
    // Create a new request multiplexer
    // Take incoming requests and dispatch them to the matching handlers
    router := mux.NewRouter()

    serverStore := repositories.NewServerStore(db)
    serverHandler := handlers.NewServerHandler(serverStore)
    userHandler := handlers.UserHandler{}

    // Register server handlers
    router.HandleFunc("/servers", serverHandler.CreateServer).Methods("POST")
    router.HandleFunc("/servers", serverHandler.ListServers).Methods("GET")
    router.HandleFunc("/servers/{id}", serverHandler.GetServer).Methods("GET")
    router.HandleFunc("/servers/{id}", serverHandler.DeleteServer).Methods("DELETE")

    // Register user handlers
    router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
    router.HandleFunc("/users", userHandler.ListUsers).Methods("GET")
    router.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
    router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")

    // Run the server
    http.ListenAndServe(":8080", router)
}
