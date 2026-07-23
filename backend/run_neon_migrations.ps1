# Script pour exécuter toutes les migrations dans le conteneur Docker PostgreSQL,
# en ciblant une base Neon distante.
# Usage:
#   Définir les variables d'environnement avant l'exécution (jamais en dur dans ce
#   fichier — audit sécurité, un mot de passe Neon réel était en clair ici) :
#     $env:NEON_DB_HOST = "..."
#     $env:NEON_DB_NAME = "..."
#     $env:NEON_DB_USER = "..."
#     $env:NEON_DB_PASSWORD = "..."
#   .\run_neon_migrations.ps1
#
# Alternative : place ces variables dans un fichier local non suivi par Git (voir
# .gitignore : .env.neon, migration.env) et charge-les avant d'appeler ce script.

$ErrorActionPreference = "Stop"

# Configuration — lue depuis l'environnement, jamais codée en dur dans ce fichier.
$CONTAINER_NAME = "mazadpay_postgres"
$DB_HOST = $env:NEON_DB_HOST
$DB_NAME = $env:NEON_DB_NAME
$DB_USER = $env:NEON_DB_USER
$DB_PASSWORD = $env:NEON_DB_PASSWORD
$MIGRATIONS_DIR = "./migrations"

if (-not $DB_HOST -or -not $DB_NAME -or -not $DB_USER -or -not $DB_PASSWORD) {
    Write-Host "Erreur: variables d'environnement manquantes." -ForegroundColor Red
    Write-Host "Définissez NEON_DB_HOST, NEON_DB_NAME, NEON_DB_USER, NEON_DB_PASSWORD avant d'exécuter ce script." -ForegroundColor Yellow
    exit 1
}

Write-Host "====================================" -ForegroundColor Cyan
Write-Host "  Exécution des migrations vers Neon" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan
Write-Host ""

# Vérifier si le conteneur est en cours d'exécution
Write-Host "Vérification du conteneur Docker..." -ForegroundColor Yellow
$containerStatus = docker ps --filter "name=$CONTAINER_NAME" --format "{{.Status}}"
if (-not $containerStatus) {
    Write-Host "Erreur: Le conteneur $CONTAINER_NAME n'est pas en cours d'exécution." -ForegroundColor Red
    Write-Host "Veuillez démarrer le conteneur avec: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

Write-Host "Conteneur $CONTAINER_NAME trouvé: $containerStatus" -ForegroundColor Green
Write-Host ""

# Attendre que PostgreSQL soit prêt
Write-Host "Attente de la disponibilité de PostgreSQL..." -ForegroundColor Yellow
$ready = $false
$maxAttempts = 30
$attempt = 0

while (-not $ready -and $attempt -lt $maxAttempts) {
    $attempt++
    try {
        docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME pg_isready -h $DB_HOST -U $DB_USER -d $DB_NAME | Out-Null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            Write-Host "PostgreSQL est prêt!" -ForegroundColor Green
        }
    } catch {
        Start-Sleep -Seconds 2
    }
}

if (-not $ready) {
    Write-Host "Erreur: PostgreSQL n'est pas prêt après $maxAttempts tentatives." -ForegroundColor Red
    exit 1
}

Write-Host ""

# Lister toutes les migrations up dans l'ordre numérique
Write-Host "Recherche des migrations..." -ForegroundColor Yellow

# Vérifier si le dossier migrations existe
if (-not (Test-Path $MIGRATIONS_DIR)) {
    Write-Host "Erreur: Le dossier $MIGRATIONS_DIR n'existe pas." -ForegroundColor Red
    exit 1
}

$migrations = Get-ChildItem -Path $MIGRATIONS_DIR -Filter "*.up.sql" | Sort-Object Name

if ($migrations.Count -eq 0) {
    Write-Host "Aucune migration trouvée dans $MIGRATIONS_DIR" -ForegroundColor Red
    exit 1
}

Write-Host "Found $($migrations.Count) migrations à exécuter:" -ForegroundColor Green
foreach ($migration in $migrations) {
    Write-Host "  - $($migration.Name)" -ForegroundColor Gray
}
Write-Host ""

# Créer une table de suivi des migrations si elle n'existe pas
Write-Host "Création de la table de suivi des migrations..." -ForegroundColor Yellow
$createTableSQL = @"
CREATE TABLE IF NOT EXISTS schema_migrations (
    version bigint PRIMARY KEY,
    dirty boolean NOT NULL DEFAULT false
);
"@
docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "$createTableSQL" | Out-Null
Write-Host "Table de suivi prête." -ForegroundColor Green
Write-Host ""

# Exécuter chaque migration
$successCount = 0
$skippedCount = 0
$errorCount = 0

foreach ($migration in $migrations) {
    if ($null -eq $migration) { continue }
    
    # Extraire le numéro de version
    if ($migration.Name -match "^(\d+)") {
        $version = [int64]$matches[1]
    } else {
        continue
    }

    Write-Host "----------------------------------------" -ForegroundColor Gray
    Write-Host "Migration: $($migration.Name)" -ForegroundColor Cyan
    
    # Vérifier si la migration a déjà été appliquée
    try {
        $checkSQL = "SELECT version FROM schema_migrations WHERE version = $version;"
        $alreadyApplied = docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME psql -h $DB_HOST -U $DB_USER -d $DB_NAME -t -c "$checkSQL"
        
        if ($alreadyApplied -and $alreadyApplied.Trim() -eq "$version") {
            Write-Host "  Déjà appliquée, ignorée." -ForegroundColor Yellow
            $skippedCount++
            continue
        }
    } catch {
        Write-Host "  Avertissement: Impossible de vérifier si déjà appliquée, continuation..." -ForegroundColor Yellow
    }
    
    # Exécuter la migration
    try {
        $tempPath = "/tmp/$($migration.Name)"
        docker cp "$($migration.FullName)" "${CONTAINER_NAME}:${tempPath}" | Out-Null
        docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $tempPath | Out-Null
        docker exec $CONTAINER_NAME rm -f $tempPath | Out-Null
        
        $insertSQL = "INSERT INTO schema_migrations (version, dirty) VALUES ($version, false);"
        docker exec -e PGPASSWORD=$DB_PASSWORD $CONTAINER_NAME psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "$insertSQL" | Out-Null
        
        Write-Host "  Appliquée avec succès!" -ForegroundColor Green
        $successCount++
    } catch {
        Write-Host "  Erreur lors de l'application: $_" -ForegroundColor Red
        $errorCount++
    }
}

Write-Host ""
Write-Host "====================================" -ForegroundColor Cyan
Write-Host "  Résumé des migrations" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan
Write-Host "Succès:     $successCount" -ForegroundColor Green
Write-Host "Ignorées:   $skippedCount" -ForegroundColor Yellow
Write-Host "Erreurs:    $errorCount" -ForegroundColor Red
Write-Host ""

if ($errorCount -gt 0) {
    Write-Host "Certaines migrations ont échoué. Veuillez vérifier les logs." -ForegroundColor Red
    exit 1
} else {
    Write-Host "Toutes les migrations ont été appliquées avec succès!" -ForegroundColor Green
    exit 0
}
