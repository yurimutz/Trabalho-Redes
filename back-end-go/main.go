package main

import (
	"log"
	"net/http"
)

func main() {
	// fmt.Println("Hello, World!")
	log.Println("Servidor de Origem pronto e escutando na porta 8081...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
