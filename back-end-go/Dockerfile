# Construtor (Baixa o Go e compila)
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY main.go .
# Compila o arquivo Go gerando um executável chamado 'servidor-origem'
RUN go build -o servidor-origem main.go

# Imagem Final (Limpa e leve)
FROM alpine:latest
WORKDIR /app
# Copia APENAS o executável pronto da etapa anterior
COPY --from=builder /app/servidor-origem .

# Cria a pasta interna onde o volume do Docker vai injetar os vídeos
RUN mkdir -p /app/videos

# Expõe a porta que definimos no main.go
EXPOSE 8080

# Comando que roda assim que o contêiner liga
ENTRYPOINT ["./servidor-origem"]