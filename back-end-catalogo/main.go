package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type VideoInfo struct {
	Name      string `json:"nome"`
	Manifest  string `json:"manifesto"`
	Thumbnail string `json:"thumbnail"`
}

func CatalogoAgregadoHandler(w http.ResponseWriter, r *http.Request) {
	// Lista de URLs internas dos seus backends dentro da rede do Docker
	backends := []string{
		"http://go_origin_1:8080/catalogo",
		"http://go_origin_2:8080/catalogo",
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var listaUnificada []VideoInfo

	// Criamos um cliente HTTP com Timeout para o agregador não ficar travado se um backend cair
	client := &http.Client{Timeout: 3 * time.Second}

	// Dispara uma Goroutine para cada backend em paralelo
	for _, url := range backends {
		wg.Add(1)

		go func(targetURL string) {
			defer wg.Done()

			res, err := client.Get(targetURL)
			if err != nil {
				fmt.Printf("Erro ao conectar no backend %s: %v\n", targetURL, err)
				return
			}
			defer res.Body.Close()

			var videosDoBackend []VideoInfo
			if err := json.NewDecoder(res.Body).Decode(&videosDoBackend); err != nil {
				fmt.Printf("Erro ao decodificar JSON de %s: %v\n", targetURL, err)
				return
			}

			// Mutex (Trava de segurança): Garante que apenas uma Goroutine altere o array por vez
			mu.Lock()
			listaUnificada = append(listaUnificada, videosDoBackend...)
			mu.Unlock()

		}(url)
	}

	// Aguarda todas as Goroutines terminarem
	wg.Wait()

	// Devolve o JSON unificado para o Nginx/Vue
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // CORS de segurança
	json.NewEncoder(w).Encode(listaUnificada)
}

func main() {
	http.HandleFunc("/catalogo", CatalogoAgregadoHandler)
	fmt.Println("BFF Agregador rodando na porta 8080...")
	http.ListenAndServe(":8080", nil)
	// log.Fatal(http.ListenAndServe(":8080", nil))
}
