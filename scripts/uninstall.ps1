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

$path = Get-UserPath
if (-not [string]::IsNullOrWhiteSpace($path)) {
    $parts = $path -split ";" | Where-Object {
        $_ -and ($_.Trim() -ne "") -and ($_.Trim().ToLower() -ne $InstallDir.ToLower())
    }
    Set-UserPath ($parts -join ";")
}

$phpHome = [Environment]::GetEnvironmentVariable("PVM_HOME", "User")
if ($phpHome -and (Test-Path $InstallDir) -and ($phpHome -eq (Join-Path $env:USERPROFILE ".pvm"))) {
    # Only remove default PVM_HOME if it matches the conventional location.
}

if (Test-Path $InstallDir) {
    Remove-Item -Recurse -Force $InstallDir
}

Write-Host "Removed pvm from user PATH and deleted $InstallDir"
Write-Host "Note: downloaded PHP versions under %USERPROFILE%\.pvm are kept."
