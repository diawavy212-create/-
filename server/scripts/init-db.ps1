param(
  [string]$HostName = "127.0.0.1",
  [int]$Port = 3306,
  [string]$User = "root",
  [string]$Password = "password",
  [string]$SchemaPath = "$PSScriptRoot\..\schema.sql"
)

$resolvedSchema = Resolve-Path -LiteralPath $SchemaPath

$mysqlCommand = Get-Command mysql -ErrorAction SilentlyContinue
if (-not $mysqlCommand) {
  $defaultMysql = "C:\Program Files\MySQL\MySQL Server 8.0\bin\mysql.exe"
  if (Test-Path -LiteralPath $defaultMysql) {
    $mysqlCommand = Get-Item -LiteralPath $defaultMysql
  } else {
    Write-Error "mysql client was not found in PATH. Install MySQL client or add mysql.exe to PATH."
    exit 1
  }
}

$mysqlArgs = @(
  "--host=$HostName",
  "--port=$Port",
  "--user=$User",
  "--password=$Password",
  "--default-character-set=utf8mb4"
)

Get-Content -LiteralPath $resolvedSchema.Path -Raw -Encoding UTF8 | & $mysqlCommand.Source @mysqlArgs

if ($LASTEXITCODE -ne 0) {
  Write-Error "Database initialization failed."
  exit $LASTEXITCODE
}

Write-Host "Database teacher_platform initialized from $resolvedSchema"
