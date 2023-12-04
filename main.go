package main

import (
	"net/http"
	"log"
)

func main(){

	router := NewRouter()
	server := http.ListenAndServe(":8080", router)
	log.Fatal(server)
}

