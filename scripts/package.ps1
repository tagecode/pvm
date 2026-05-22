param(
    [Parameter(Mandatory = $true)]
    [string]$BinaryPath,

    [Parameter(Mandatory = $true)]
    [string]$OutputDir,

    [Parameter(Mandatory = $true)]
    [string]$Version,

    [Parameter(Mandatory = $true)]
    [ValidateSet("amd64", "386", "x64", "x86")]
    [string]$Arch,

    [string]$StandaloneExePath = ""
)

$ErrorActionPreference = "Stop"
. "$PSScriptRoot/naming.ps1"

$archLabel = ConvertTo-PvmArchLabel -Arch $Arch
$versionTag = Get-PvmVersionTag -Version $Version
$packageName = Get-PvmArtifactBaseName -Version $Version -Arch $Arch
$stage = Join-Path $OutputDir $packageName
$scriptsDir = $PSScriptRoot

if (Test-Path $stage) {
    Remove-Item -Recurse -Force $stage
}
New-Item -ItemType Directory -Path $stage | Out-Null

Copy-Item $BinaryPath (Join-Path $stage "pvm.exe")
Copy-Item (Join-Path $scriptsDir "install.ps1") (Join-Path $stage "install.ps1")
Copy-Item (Join-Path $scriptsDir "uninstall.ps1") (Join-Path $stage "uninstall.ps1")

@"
PVM $versionTag for Windows $archLabel

Files:
  pvm.exe       - PVM CLI binary
  install.ps1   - Install to user PATH (recommended)
  uninstall.ps1 - Remove from user PATH

Quick start:
  1. Open PowerShell in this folder
  2. Run: .\install.ps1
  3. Open a new terminal and run: pvm version

Manual use:
  Copy pvm.exe anywhere and run it directly without installing.

Docs: https://github.com/tagecode/pvm
"@ | Set-Content -Path (Join-Path $stage "README.txt") -Encoding UTF8

$zipPath = Join-Path $OutputDir "$packageName.zip"
if (Test-Path $zipPath) {
    Remove-Item -Force $zipPath
}
Compress-Archive -Path (Join-Path $stage "*") -DestinationPath $zipPath

$hash = Get-FileHash -Path $zipPath -Algorithm SHA256
$exeHash = Get-FileHash -Path (Join-Path $stage "pvm.exe") -Algorithm SHA256

$checksumFile = Join-Path $OutputDir "checksums-$archLabel.txt"
$checksumLines = @(
    "$($hash.Hash.ToLower())  $packageName.zip"
    "$($exeHash.Hash.ToLower())  $packageName/pvm.exe"
)
if ($StandaloneExePath -and (Test-Path $StandaloneExePath)) {
    $standaloneHash = Get-FileHash -Path $StandaloneExePath -Algorithm SHA256
    $checksumLines += "$($standaloneHash.Hash.ToLower())  $(Split-Path -Leaf $StandaloneExePath)"
}
$checksumLines | Set-Content -Path $checksumFile -Encoding ASCII

Write-Output $zipPath
