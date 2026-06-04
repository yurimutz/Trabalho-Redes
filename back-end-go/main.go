package main

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

func main() {
	// fmt.Println("Hello, World!")

	pastaVideos := "/app/videos"

	// http.HandleFunc("/teste/", func(w http.ResponseWriter, r *http.Request) {
	// 	if(r.Method != "GET"){
	// 	}
	// 	//Teste de retornos
	// 	io.WriteString(w, "Hello from a HandleFunc #2 %s!\n", r.URL)
	// 	fmt.Fprintf(w, "Hello from a HandleFunc #2 %s!\n", r.URL)
	// })

	http.HandleFunc("/videos/", func(w http.ResponseWriter, r *http.Request) {
		nomeArquivo := strings.TrimPrefix(r.URL.Path, "/videos/")
		caminhoFisico := filepath.Join(pastaVideos, nomeArquivo)
		http.ServeFile(w, r, caminhoFisico)

	})

	http.HandleFunc("/manifesto/", func(w http.ResponseWriter, r *http.Request) {
		//fmt.Fprintf(w, "Hello from a HandleFunc #2 %s!\n", r.URL)
		nomeArquivo := strings.TrimPrefix(r.URL.Path, "/manifesto/")
		caminhoFisico := filepath.Join(pastaVideos, nomeArquivo)
		http.ServeFile(w, r, caminhoFisico)
	})

	log.Println("Servidor de Origem pronto e escutando na porta 8082...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
