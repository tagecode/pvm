param(
    [string]$InstallDir = $(Join-Path $env:LOCALAPPDATA "Programs\pvm")
)

$ErrorActionPreference = "Stop"

function Get-UserPath {
    return [Environment]::GetEnvironmentVariable("Path", "User")
}

function Set-UserPath([string]$Path) {
    [Environment]::SetEnvironmentVariable("Path", $Path, "User")
}

function Add-UserPathEntry([string]$PathValue, [string]$Entry) {
    $Entry = $Entry.Trim()
    if ([string]::IsNullOrWhiteSpace($Entry)) {
        return $PathValue
    }
    $parts = @()
    if (-not [string]::IsNullOrWhiteSpace($PathValue)) {
        $parts = $PathValue -split ";" | Where-Object { $_ -and ($_.Trim() -ne "") -and ($_.Trim().ToLower() -ne $Entry.ToLower()) }
    }
    if ($parts.Count -eq 0) {
        return $Entry
    }
    return $Entry + ";" + ($parts -join ";")
}

$sourceDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$sourceExe = Join-Path $sourceDir "pvm.exe"
if (-not (Test-Path $sourceExe)) {
    throw "pvm.exe not found next to install.ps1"
}

New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Copy-Item $sourceExe (Join-Path $InstallDir "pvm.exe") -Force

$pvmHome = Join-Path $env:USERPROFILE ".pvm"
$current = Join-Path $pvmHome "current"
New-Item -ItemType Directory -Path $pvmHome -Force | Out-Null

$path = Get-UserPath
$path = Add-UserPathEntry $path $current
$path = Add-UserPathEntry $path $InstallDir
Set-UserPath $path

[Environment]::SetEnvironmentVariable("PVM_HOME", $pvmHome, "User")
[Environment]::SetEnvironmentVariable("PHP_HOME", $current, "User")

Write-Host "Installed pvm to $InstallDir"
Write-Host "PVM_HOME=$pvmHome"
Write-Host "PHP_HOME=$current"
Write-Host "Run 'pvm install 8.3' then use 'php -v' in a new terminal (or after first use in current terminal)."
