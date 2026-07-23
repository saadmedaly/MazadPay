# migrate_neon.ps1
# Run all migrations on Neon PostgreSQL in correct order (auto-discovered from
# backend/migrations/*.up.sql, sorted by filename).
# Usage: .\migrate_neon.ps1 -NeonUrl "postgresql://user:pass@host/db?sslmode=require"
# Resume: .\migrate_neon.ps1 -NeonUrl "..." -ResumeFrom "000004"
# Help:   .\migrate_neon.ps1 -Help
#
# -NeonUrl contient des identifiants sensibles : ne jamais le coder en dur dans un
# script ou le committer. Le passer uniquement en ligne de commande ou via une
# variable d'environnement locale non suivie par Git (voir .gitignore : .env.neon,
# migration.env).

param(
    [Parameter(Mandatory=$false)]
    [switch]$Help,

    [Parameter(Mandatory=$false)]
    [string]$NeonUrl,

    [Parameter(Mandatory=$false)]
    [string]$ResumeFrom = ""
)

if ($Help) {
    Write-Host "migrate_neon.ps1 - Applique les migrations MazadPay sur une base Neon PostgreSQL"
    Write-Host ""
    Write-Host "Usage:"
    Write-Host "  .\migrate_neon.ps1 -NeonUrl <connection-string> [-ResumeFrom <prefix>]"
    Write-Host "  .\migrate_neon.ps1 -Help"
    Write-Host ""
    Write-Host "Parametres:"
    Write-Host "  -NeonUrl     Chaine de connexion Postgres complete (ex: postgresql://user:pass@host/db?sslmode=require)."
    Write-Host "               Jamais codee en dur ici -- passer en ligne de commande ou via variable d'environnement locale."
    Write-Host "  -ResumeFrom  Prefixe de migration (ex: '000039') a partir duquel reprendre, pour une base non vide."
    Write-Host "  -Help        Affiche cette aide et quitte sans se connecter a une base de donnees."
    Write-Host ""
    Write-Host "Les migrations sont decouvertes automatiquement depuis backend/migrations/*.up.sql,"
    Write-Host "triees par nom de fichier -- aucune liste codee en dur a maintenir."
    exit 0
}

if (-not $NeonUrl) {
    Write-Host "ERROR: -NeonUrl is required (or use -Help for usage)." -ForegroundColor Red
    exit 1
}

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$MigrationsDir = Join-Path $ScriptDir "migrations"

# Auto-discovery : toutes les migrations *.up.sql du dossier, triées par nom (donc par
# préfixe numérique) — remplace l'ancienne liste codée en dur qui s'arrêtait à 000038
# et avait été oubliée lors de l'ajout de 000039/000040/000041 (Migration Tooling
# Security Hardening). Évite que ce script ne devienne silencieusement obsolète à
# chaque nouvelle migration ajoutée au dossier.
#
# Note sur les préfixes en doublon (ex: 000003_add_admin / 000003_create_requests) :
# le tri alphabétique sur le nom complet du fichier reste stable et déterministe
# (ordre alphabétique du descriptif après le préfixe), identique au comportement de
# run_migrations.ps1 qui utilise la même logique Sort-Object Name.
$migrations = (Get-ChildItem -Path $MigrationsDir -Filter "*.up.sql" | Sort-Object Name).Name
if (-not $migrations -or $migrations.Count -eq 0) {
    # 000002_seed.sql n'a pas de suffixe .up.sql (cas historique isolé) — inclus
    # explicitement s'il existe, pour ne pas le perdre silencieusement.
    Write-Host "ERROR: No *.up.sql migrations found in $MigrationsDir" -ForegroundColor Red
    exit 1
}
$seedFile = "000002_seed.sql"
if ((Test-Path (Join-Path $MigrationsDir $seedFile)) -and ($migrations -notcontains $seedFile)) {
    $migrations = @($migrations | Where-Object { $_ -lt $seedFile }) + @($seedFile) + @($migrations | Where-Object { $_ -gt $seedFile })
}

# -- ResumeFrom mode: skip empty-DB check, filter migration list --
if ($ResumeFrom -ne "") {
    Write-Host ""
    Write-Host "[RESUME] Starting from prefix: $ResumeFrom" -ForegroundColor Yellow
    Write-Host "[RESUME] Skipping empty-database check." -ForegroundColor Yellow

    # Find index of first migration whose name starts with ResumeFrom
    $startIndex = -1
    for ($i = 0; $i -lt $migrations.Count; $i++) {
        if ($migrations[$i].StartsWith($ResumeFrom)) {
            $startIndex = $i
            break
        }
    }

    if ($startIndex -eq -1) {
        Write-Host "ERROR: No migration found with prefix '$ResumeFrom'." -ForegroundColor Red
        exit 1
    }

    Write-Host "[RESUME] Found start at index ${startIndex}: $($migrations[$startIndex])" -ForegroundColor Yellow
    $migrations = $migrations[$startIndex..($migrations.Count - 1)]

} else {
    # Step 1: Check if DB is empty
    Write-Host ""
    Write-Host "[1/3] Checking if Neon database is empty..." -ForegroundColor Cyan
    $checkResult = & psql $NeonUrl -t -A -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>&1
    if ($LASTEXITCODE -ne 0) {
        Write-Host "ERROR: Cannot connect to Neon. Check your URL." -ForegroundColor Red
        Write-Host $checkResult
        exit 1
    }

    $tableCount = [int](($checkResult | Select-Object -Last 1).Trim())
    Write-Host "Tables found: $tableCount"

    if ($tableCount -gt 0) {
        Write-Host ""
        Write-Host "WARNING: Database already has $tableCount tables." -ForegroundColor Yellow
        Write-Host "Existing tables:" -ForegroundColor Yellow
        & psql $NeonUrl -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"
        Write-Host ""
        Write-Host "Aborted. Use -ResumeFrom to continue from a specific migration." -ForegroundColor Red
        exit 1
    }

    Write-Host "Database is empty. Proceeding with migrations." -ForegroundColor Green
}

# Step 2: Run migrations
Write-Host ""
Write-Host "[2/3] Running $($migrations.Count) migrations..." -ForegroundColor Cyan
$success = 0

foreach ($file in $migrations) {
    $path = Join-Path $MigrationsDir $file
    if (-not (Test-Path $path)) {
        Write-Host "  SKIP (not found): $file" -ForegroundColor Yellow
        continue
    }

    Write-Host "  Running: $file" -NoNewline

    # Temporarily suspend Stop preference so stderr from psql does not throw
    $oldEap = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $output = & psql $NeonUrl -v ON_ERROR_STOP=1 -f $path 2>&1
    $exitCode = $LASTEXITCODE
    $ErrorActionPreference = $oldEap

    # Print NOTICE/WARNING lines as info only
    $noticeLines = $output | Where-Object { $_ -match '^(NOTICE|WARNING)' }
    if ($noticeLines) {
        Write-Host ""
        foreach ($n in $noticeLines) { Write-Host "    [NOTICE] $n" -ForegroundColor DarkYellow }
        Write-Host "  Running: $file" -NoNewline
    }

    if ($exitCode -ne 0) {
        Write-Host " FAILED" -ForegroundColor Red
        Write-Host ""
        Write-Host "=== Migration failed at: $file ===" -ForegroundColor Red
        Write-Host "=== File path: $path ===" -ForegroundColor Red
        Write-Host "=== Exit code: $exitCode ===" -ForegroundColor Red
        Write-Host "=== Output ===" -ForegroundColor Red
        $output | ForEach-Object { Write-Host "$_" -ForegroundColor Red }
        Write-Host "=== End of output ===" -ForegroundColor Red
        Write-Host "Stopping. Fix the error above before continuing." -ForegroundColor Red
        exit 1
    }
    Write-Host " OK" -ForegroundColor Green
    $success++
}

# Step 3: Verify essential tables
Write-Host ""
Write-Host "[3/3] Verifying essential tables..." -ForegroundColor Cyan
$essential = @("users","auctions","categories","bids","countries","locations","wallets","notifications")
$allOk = $true

foreach ($table in $essential) {
    $exists = & psql $NeonUrl -t -A -c "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='$table');" 2>&1
    if (($exists | Select-Object -Last 1).Trim() -eq "t") {
        Write-Host "  [OK] $table" -ForegroundColor Green
    } else {
        Write-Host "  [MISSING] $table" -ForegroundColor Red
        $allOk = $false
    }
}

# Final count
$finalCount = & psql $NeonUrl -t -A -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>&1

Write-Host ""
Write-Host "==============================" -ForegroundColor Cyan
Write-Host "Migrations run : $success" -ForegroundColor Green
Write-Host "Total tables   : $(($finalCount | Select-Object -Last 1).Trim())" -ForegroundColor Green

if ($allOk) {
    Write-Host "Essential tables: ALL OK" -ForegroundColor Green
    Write-Host ""
    Write-Host "Neon database is ready for production." -ForegroundColor Green
} else {
    Write-Host "Some essential tables are MISSING - check errors above." -ForegroundColor Red
    exit 1
}
