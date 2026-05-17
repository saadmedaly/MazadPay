# Stage 1 — Builder
FROM golang:1.25-alpine AS builder

# Installation des certificats CA et de tzdata (pour la gestion des timezones)
RUN apk --no-cache add ca-certificates tzdata

# Configuration de l'environnement de build pour un binaire statique minimal
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

# Optimisation du cache des dépendances (layer caching)
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste du code
COPY . .

# Compilation ultra-optimisée
# -trimpath : retire les chemins locaux du système de build du binaire
# -ldflags="-s -w" : supprime la table des symboles et les infos de debug (réduit drastiquement la taille)
RUN go build -trimpath -ldflags="-s -w" -o /app/server ./cmd/server

# ==============================================================================
# Stage 2 — Final (Image vide ultra-légère)
# ==============================================================================
FROM scratch

# Nécessaire pour faire des requêtes TLS/HTTPS sortantes (Neon, S3/R2, Stripe) et pour la gestion du temps
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copier uniquement le binaire compilé et les configurations nécessaires
COPY --from=builder /app/server /server
COPY --from=builder /app/configs /configs

# Exposition du port
EXPOSE 8082

# Démarrage
ENTRYPOINT ["/server"]
