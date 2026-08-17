param([string]$DatabaseUrl = $env:DATABASE_URL)
if (-not $DatabaseUrl) { throw 'DATABASE_URL is required' }
psql $DatabaseUrl -v ON_ERROR_STOP=1 -f (Join-Path $PSScriptRoot '..\migrations\010_review_schema.sql')
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
psql $DatabaseUrl -v ON_ERROR_STOP=1 -f (Join-Path $PSScriptRoot '..\migrations\020_demo_review.sql')
