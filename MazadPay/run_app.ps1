#!/usr/bin/env pwsh
# Script pour lancer l'application MazadPay Flutter

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "    MazadPay Flutter - Build Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Étape 1: Générer les localisations
Write-Host "[1/4] Génération des fichiers de localisation..." -ForegroundColor Yellow
try {
    flutter gen-l10n
    Write-Host "      ✓ Localisations générées" -ForegroundColor Green
} catch {
    Write-Host "      ⚠ Erreur lors de la génération (peut être ignorée)" -ForegroundColor Yellow
}

# Étape 2: Nettoyer
Write-Host "[2/4] Nettoyage du projet..." -ForegroundColor Yellow
flutter clean
Write-Host "      ✓ Projet nettoyé" -ForegroundColor Green

# Étape 3: Récupérer les dépendances
Write-Host "[3/4] Récupération des dépendances..." -ForegroundColor Yellow
flutter pub get
Write-Host "      ✓ Dépendances récupérées" -ForegroundColor Green

# Étape 4: Lancer l'application
Write-Host "[4/4] Lancement de l'application..." -ForegroundColor Yellow
Write-Host ""
Write-Host "Options disponibles:" -ForegroundColor Cyan
Write-Host "  1. Chrome (Web)"
Write-Host "  2. Windows (Desktop)"
Write-Host "  3. Android (si connecté)"
Write-Host ""

$choice = Read-Host "Choisissez une option (1-3, défaut: 1)"

switch ($choice) {
    "2" { flutter run -d windows }
    "3" { flutter run -d android }
    default { flutter run -d chrome }
}
