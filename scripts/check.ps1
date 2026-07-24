param(
  [switch]$SkipAdminBuild
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot

function Run-Step {
  param(
    [string]$Name,
    [scriptblock]$Action
  )

  Write-Host ""
  Write-Host "==> $Name"
  & $Action
  Write-Host "OK: $Name"
}

Run-Step "Go packages compile" {
  Push-Location (Join-Path $Root "server")
  try {
    go test ./...
  } finally {
    Pop-Location
  }
}

if (-not $SkipAdminBuild) {
  Run-Step "Admin production build" {
    Push-Location (Join-Path $Root "admin")
    try {
      $vite = Join-Path (Get-Location) "node_modules\.bin\vite.cmd"
      if (-not (Test-Path $vite)) {
        throw "admin dependencies are missing. Run pnpm install in admin first."
      }
      & $vite build
    } finally {
      Pop-Location
    }
  }
}

Run-Step "Miniprogram JSON, page files, icons, and JS syntax" {
  Push-Location $Root
  try {
    node -e @"
const fs = require('fs');
const { spawnSync } = require('child_process');

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

const app = readJSON('miniprogram/app.json');
readJSON('miniprogram/project.config.json');

for (const page of app.pages) {
  for (const ext of ['.js', '.json', '.wxml', '.wxss']) {
    const file = 'miniprogram/' + page + ext;
    if (!fs.existsSync(file)) throw new Error('Missing page file: ' + file);
    if (ext === '.json') readJSON(file);
  }
}

for (const item of app.tabBar.list || []) {
  for (const key of ['iconPath', 'selectedIconPath']) {
    const file = 'miniprogram/' + item[key];
    if (!fs.existsSync(file)) throw new Error('Missing tabBar icon: ' + file);
  }
}

function walk(dir, result = []) {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = dir + '/' + entry.name;
    if (entry.isDirectory()) walk(full, result);
    else if (entry.isFile() && full.endsWith('.js')) result.push(full);
  }
  return result;
}

for (const file of walk('miniprogram')) {
  const checked = spawnSync(process.execPath, ['--check', file], { stdio: 'inherit' });
  if (checked.status !== 0) process.exit(checked.status || 1);
}

console.log('Miniprogram checks passed');
"@
  } finally {
    Pop-Location
  }
}

Write-Host ""
Write-Host "All automated checks passed."
