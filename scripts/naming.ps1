function ConvertTo-PvmArchLabel {
    param(
        [Parameter(Mandatory = $true)]
        [ValidateSet("amd64", "386", "x64", "x86")]
        [string]$Arch
    )

    switch ($Arch) {
        "amd64" { return "x64" }
        "386" { return "x86" }
        "x64" { return "x64" }
        "x86" { return "x86" }
    }
}

function ConvertTo-PvmGoArch {
    param(
        [Parameter(Mandatory = $true)]
        [ValidateSet("amd64", "386", "x64", "x86")]
        [string]$Arch
    )

    switch ($Arch) {
        "amd64" { return "amd64" }
        "386" { return "386" }
        "x64" { return "amd64" }
        "x86" { return "386" }
    }
}

function Get-PvmVersionTag {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Version
    )

    $Version = $Version.Trim().TrimStart("v")
    if ([string]::IsNullOrWhiteSpace($Version)) {
        throw "version is empty"
    }
    return "v$Version"
}

function Get-PvmArtifactBaseName {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Version,

        [Parameter(Mandatory = $true)]
        [ValidateSet("amd64", "386", "x64", "x86")]
        [string]$Arch
    )

    $versionTag = Get-PvmVersionTag -Version $Version
    $archLabel = ConvertTo-PvmArchLabel -Arch $Arch
    return "pvm-$versionTag-windows-$archLabel"
}
