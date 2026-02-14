param(
  [switch]$Build
)

$ErrorActionPreference = 'Stop'

$here = Split-Path -Parent $MyInvocation.MyCommand.Path
$root = Split-Path -Parent $here

Push-Location $root
try {
  if ($Build) {
    docker compose up --build -d
  } else {
    docker compose up -d
  }
} finally {
  Pop-Location
}
