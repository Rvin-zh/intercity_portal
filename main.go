package main

import (
        "fmt"
        "log"
        "net/http"
)

func main() {
        // Set up file server for static files
        fs := http.FileServer(http.Dir("static"))
        http.Handle("/static/", http.StripPrefix("/static/", fs))

        // Set up routes
        http.HandleFunc("/", indexHandler)
        http.HandleFunc("/login", loginHandler)
        http.HandleFunc("/auth", authHandler)
        http.HandleFunc("/forgot", forgotHandler)
        http.HandleFunc("/reset-password", resetPasswordHandler)

        // Start the server
        port := "5000"
        addr := "0.0.0.0:" + port
        fmt.Printf("Server started on http://%s\n", addr)
        log.Fatal(http.ListenAndServe(addr, nil))
}
