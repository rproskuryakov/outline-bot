package handlers

import (
    "net/http"
)


type ServerHandler struct{}

func (h *ServerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *ServerHandler) CreateServer(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *ServerHandler) DeleteServer(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *ServerHandler) GetServer(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *ServerHandler) ListServers(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}


type UserHandler struct {}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("This is my home page"))
}


