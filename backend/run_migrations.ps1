# Script pour exécuter toutes les migrations dans le conteneur Docker PostgreSQL
# Usage: .\run_migrations.ps1

$ErrorActionPreference = "Stop"

# Configuration
$CONTAINER_NAME = "mazadpay_postgres"
$DB_NAME = "mazadpay"
$DB_USER = "mazadpay"
$DB_PASSWORD = "mazadpay_secret"
$MIGRATIONS_DIR = "./migrations"

Write-Host "====================================" -ForegroundColor Cyan
Write-Host "  Exécution des migrations MazadPay" -ForegroundColor Cyan
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
        docker exec $CONTAINER_NAME pg_isready -U $DB_USER -d $DB_NAME | Out-Null
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
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
"@
docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c "$createTableSQL" | Out-Null
Write-Host "Table de suivi prête." -ForegroundColor Green
Write-Host ""

# Exécuter chaque migration
$successCount = 0
$skippedCount = 0
$errorCount = 0

foreach ($migration in $migrations) {
    # Vérifier que l'objet migration n'est pas null
    if ($null -eq $migration) {
        Write-Host "  Erreur: Migration null trouvée" -ForegroundColor Red
        continue
    }
    
    # Vérifier que FullName existe
    if (-not $migration.FullName) {
        Write-Host "  Erreur: Impossible de lire le fichier pour $($migration.Name)" -ForegroundColor Red
        continue
    }
    
    $version = $migration.Name -replace '\.up\.sql$', ''
    Write-Host "----------------------------------------" -ForegroundColor Gray
    Write-Host "Migration: $($migration.Name)" -ForegroundColor Cyan
    
    # Vérifier si la migration a déjà été appliquée
    try {
        $checkSQL = "SELECT version FROM schema_migrations WHERE version = '$version';"
        $alreadyApplied = docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -t -c "$checkSQL"
        
        if ($alreadyApplied.Trim() -eq $version) {
            Write-Host "  Déjà appliquée, ignorée." -ForegroundColor Yellow
            $skippedCount++
            continue
        }
    } catch {
        Write-Host "  Avertissement: Impossible de vérifier si déjà appliquée, continuation..." -ForegroundColor Yellow
    }
    
    # Lire le contenu de la migration
    $migrationContent = Get-Content $migration.FullName -Raw -Encoding UTF8
    
    # Exécuter la migration
    try {
        # Copier le fichier de migration dans le conteneur
        $tempPath = "/tmp/$($migration.Name)"
        docker cp "$($migration.FullName)" "${CONTAINER_NAME}:${tempPath}" | Out-Null
        
        # Exécuter la migration depuis le conteneur
        docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -f $tempPath | Out-Null
        
        # Nettoyer le fichier temporaire
        docker exec $CONTAINER_NAME rm -f $tempPath | Out-Null
        
        # Marquer la migration comme appliquée
        $insertSQL = "INSERT INTO schema_migrations (version) VALUES ('$version');"
        docker exec $CONTAINER_NAME psql -U $DB_USER -d $DB_NAME -c "$insertSQL" | Out-Null
        
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
