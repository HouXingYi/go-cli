$ErrorActionPreference = "Stop"

$AppName = "zn-cli"
$Entry = "./cmd/ziniao"
$DistDir = "dist"

$Targets = @(
    @{ GOOS = "linux"; GOARCH = "amd64" },
    @{ GOOS = "linux"; GOARCH = "arm64" },
    @{ GOOS = "darwin"; GOARCH = "amd64" },
    @{ GOOS = "darwin"; GOARCH = "arm64" },
    @{ GOOS = "windows"; GOARCH = "amd64" }
)

if (Test-Path $DistDir) {
    Remove-Item -Recurse -Force $DistDir
}

New-Item -ItemType Directory -Force $DistDir | Out-Null

foreach ($Target in $Targets) {
    $Goos = $Target.GOOS
    $Goarch = $Target.GOARCH
    $Output = Join-Path $DistDir "$AppName-$Goos-$Goarch"

    if ($Goos -eq "windows") {
        $Output = "$Output.exe"
    }

    Write-Host "Building $Output"
    $env:GOOS = $Goos
    $env:GOARCH = $Goarch
    $env:CGO_ENABLED = "0"

    go build -o $Output $Entry
}

Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue

Write-Host "Build artifacts are in $DistDir/"
