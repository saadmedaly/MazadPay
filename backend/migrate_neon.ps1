# migrate_neon.ps1
# Run all migrations on Neon PostgreSQL in correct order
# Usage: .\migrate_neon.ps1 -NeonUrl "postgresql://user:pass@host/db?sslmode=require"
# Resume: .\migrate_neon.ps1 -NeonUrl "..." -ResumeFrom "000004"

param(
    [Parameter(Mandatory=$true)]
    [string]$NeonUrl,

    [Parameter(Mandatory=$false)]
    [string]$ResumeFrom = ""
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$MigrationsDir = Join-Path $ScriptDir "migrations"

# Ordered migration list -- duplicates resolved by dependency order
$migrations = @(
    "000001_init.up.sql",
    "000002_seed.sql",
    "000003_add_admin.up.sql",
    "000003_create_requests_tables.up.sql",
    "000004_auction_translations.up.sql",
    "000005_auto_lot_number.up.sql",
    "000006_add_name_en_to_categories.up.sql",
    "000007_update_locations_schema.up.sql",
    "000008_admin_invitations.up.sql",
    "000009_allow_web_platform.up.sql",
    "000010_add_updated_at_push_tokens.up.sql",
    "000011_increase_device_id_length.up.sql",
    "000012_audit_logs.up.sql",
    "000013_system_settings.up.sql",
    "000014_add_title_en_to_banners.up.sql",
    "000015_fix_banners_title_en.up.sql",
    "000016_add_ratings_table.up.sql",
    "000017_add_is_super_admin.up.sql",
    "000018_seed_super_admin.up.sql",
    "000019_fix_super_admin_password.up.sql",
    "000020_password_reset_tracking.up.sql",
    "000021_add_countries_support.up.sql",
    "000022_replace_termii_with_twilio.up.sql",
    "000023_add_user_optional_fields.up.sql",
    "000024_add_auctions_fields.up.sql",
    "000025_improve_categories.up.sql",
    "000026_add_transactions_fields.up.sql",
    "000027_improve_service_requests.up.sql",
    "000028_optimize_bids.up.sql",
    "000029_improve_delivery_timeline.up.sql",
    "000030_improve_notifications.up.sql",
    "000031_create_new_tables.up.sql",
    "000032_add_quantity.up.sql",
    "000032_add_updated_at_notifications.up.sql",
    "000033_create_conversations.up.sql",
    "000033_add_quantity_requests.up.sql",
    "000034_fix_conversations_view.up.sql",
    "000035_add_country_code.up.sql",
    "000036_add_app_features.up.sql",
    "000037_add_descriptions_to_banner_requests.up.sql",
    "000038_expand_notification_types.up.sql"
)

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
