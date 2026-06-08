package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type VideoInfo struct {
	Name      string `json:"nome"`
	Manifest  string `json:"manifesto"`
	Thumbnail string `json:"thumbnail"`
}

func buscarThumbnail(caminhoDaPasta string, nomeDaPasta string) string {
	arquivos, err := os.ReadDir(caminhoDaPasta)
	if err != nil {
		return ""
	}

	for _, arq := range arquivos {
		// Ignora subpastas e checa se o nome termina com .png
		if !arq.IsDir() && (strings.HasSuffix(strings.ToLower(arq.Name()), ".jpg") || strings.HasSuffix(strings.ToLower(arq.Name()), ".png")) {
			// Monta a URL que o Vue usará na tag <img src="...">
			return fmt.Sprintf("/videos/%s/%s", nomeDaPasta, arq.Name())
		}
	}

	return "" // Retorna vazio caso o vídeo ainda não tenha thumbnail
}

func ListaVideosHandler(w http.ResponseWriter, r *http.Request) {
	// Caminho onde o volume do Docker está montado no Go
	baseDir := "/app/videos"

	// Lê tudo o que está na raiz da pasta de vídeos
	entradas, err := os.ReadDir(baseDir)
	if err != nil {
		http.Error(w, "Erro ao ler o diretório de vídeos", http.StatusInternalServerError)
		return
	}

	var listaVideos []VideoInfo

	for _, entrada := range entradas {
		if entrada.IsDir() {
			nomeDaPasta := entrada.Name()
			caminhoDaPasta := filepath.Join(baseDir, nomeDaPasta)

			// O nome formatado (substitui underlines por espaços para ficar bonito no front)
			tituloLimpo := strings.ReplaceAll(nomeDaPasta, "_", " ")
			tituloLimpo = strings.Title(tituloLimpo) // Ex: "meu_video" vira "Meu Video"

			// A URL do manifesto (baseada na regra do seu Nginx/Vite)
			urlManifesto := fmt.Sprintf("/videos/%s/manifesto.mpd", nomeDaPasta)

			// Procura a imagem da Thumbnail (.png) dinamicamente
			urlThumbnail := buscarThumbnail(caminhoDaPasta, nomeDaPasta)

			// Adiciona ao array
			listaVideos = append(listaVideos, VideoInfo{
				Name:      tituloLimpo,
				Manifest:  urlManifesto,
				Thumbnail: urlThumbnail,
			})
		}
	}

	// Responde com o JSON formatado
	w.Header().Set("Content-Type", "application/json")
	// Evita problemas de CORS caso não esteja usando o Proxy do Vite no momento

	json.NewEncoder(w).Encode(listaVideos)
}

func main() {
	// fmt.Println("Hello, World!")

	pastaVideos := "/app/videos"

	http.HandleFunc("/videos/", func(w http.ResponseWriter, r *http.Request) {
		nomeArquivo := strings.TrimPrefix(r.URL.Path, "/videos/")
		caminhoFisico := filepath.Join(pastaVideos, nomeArquivo)
		http.ServeFile(w, r, caminhoFisico)

	})

	http.HandleFunc("/manifesto/", func(w http.ResponseWriter, r *http.Request) {
		nomeArquivo := strings.TrimPrefix(r.URL.Path, "/manifesto/")
		caminhoFisico := filepath.Join(pastaVideos, nomeArquivo)
		http.ServeFile(w, r, caminhoFisico)
	})

	http.HandleFunc("/catalogo", ListaVideosHandler)

	log.Println("Servidor de Origem pronto e escutando na porta 8082...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
