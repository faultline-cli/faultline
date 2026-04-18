package main

import (
	"fmt"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://localhost:8080/health")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_ = resp.Body.Close()
	w.WriteHeader(http.StatusOK)
}

func ProcessRequest(w http.ResponseWriter, r *http.Request) {
	db, err := getDB()
	if err != nil {
		panic("database unavailable")
	}
	_, _ = db.Exec("INSERT INTO logs DEFAULT VALUES")
	fmt.Fprintf(w, "ok")
}

func getDB() error {
	return nil
}
