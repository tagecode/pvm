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

    [string]$WixSource = (Join-Path $PSScriptRoot "..\installer\pvm.wxs")
)

$ErrorActionPreference = "Stop"
. "$PSScriptRoot/naming.ps1"

if (-not (Test-Path $BinaryPath)) {
    throw "binary not found: $BinaryPath"
}
if (-not (Test-Path $WixSource)) {
    throw "wix source not found: $WixSource"
}

$archLabel = ConvertTo-PvmArchLabel -Arch $Arch
$versionTag = Get-PvmVersionTag -Version $Version
$artifactBase = Get-PvmArtifactBaseName -Version $Version -Arch $Arch
$platform = if ($archLabel -eq "x86") { "x86" } else { "x64" }

# WiX Product/@Version accepts major.minor.build
$productVersion = ($Version.Trim().TrimStart("v") -split "-")[0]
if ($productVersion -notmatch "^\d+\.\d+\.\d+$") {
    throw "invalid product version for MSI: $productVersion"
}

$candle = Get-Command candle.exe -ErrorAction SilentlyContinue
$light = Get-Command light.exe -ErrorAction SilentlyContinue
if (-not $candle -or -not $light) {
    throw "WiX Toolset not found. Install wixtoolset and ensure candle.exe/light.exe are on PATH."
}

New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null
$workDir = Join-Path $OutputDir "wix-$artifactBase"
if (Test-Path $workDir) {
    Remove-Item -Recurse -Force $workDir
}
New-Item -ItemType Directory -Path $workDir | Out-Null

$wixObj = Join-Path $workDir "pvm.wixobj"
$msiPath = Join-Path $OutputDir "$artifactBase.msi"

$candleArgs = @(
    "-nologo"
    "-out", $wixObj
    "-dProductVersion=$productVersion"
    "-dVersionTag=$versionTag"
    "-dArchLabel=$archLabel"
    "-dPlatform=$platform"
    "-dPvmBinary=$BinaryPath"
    $WixSource
)
& $candle @candleArgs
if ($LASTEXITCODE -ne 0) {
    throw "candle failed with exit code $LASTEXITCODE"
}

$lightArgs = @(
    "-nologo"
    "-out", $msiPath
    $wixObj
)
& $light @lightArgs
if ($LASTEXITCODE -ne 0) {
    throw "light failed with exit code $LASTEXITCODE"
}

$msiHash = Get-FileHash -Path $msiPath -Algorithm SHA256
$checksumFile = Join-Path $OutputDir "checksums-$archLabel.txt"
$line = "$($msiHash.Hash.ToLower())  $artifactBase.msi"
if (Test-Path $checksumFile) {
    Add-Content -Path $checksumFile -Value $line -Encoding ASCII
} else {
    Set-Content -Path $checksumFile -Value $line -Encoding ASCII
}

Write-Output $msiPath
