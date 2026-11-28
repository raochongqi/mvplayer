<#
.SYNOPSIS
    Build the Windows exe (+ optional NSIS installer) inside Docker.

.DESCRIPTION
    Cross-compiles mvplayer.exe with mingw-w64 inside the official golang image,
    so no MSVC/mingw toolchain is needed on the host. The container has no WSL
    dependency beyond whatever Docker Desktop itself uses.

    Outputs land in ./build (gitignored):
        build/mvplayer.exe      GUI exe (-H windowsgui)
        build/libmpv-2.dll      ~116 MiB, fetched once into third_party/mpv
        build/mvPlayer-0.1.0-setup.exe   only with -Installer

.EXAMPLE
    ./scripts/build-windows-docker.ps1
    ./scripts/build-windows-docker.ps1 -Installer
    ./scripts/build-windows-docker.ps1 -Image golang:1.27-bookworm -AptMirror https://mirrors.tuna.tsinghua.edu.cn/debian
#>
[CmdletBinding()]
param(
    [string]$Image    = 'golang:1.27.4-bookworm',
    [switch]$Installer,
    [string]$AptMirror    = 'https://mirrors.aliyun.com/debian',
    [string]$AptMirrorSec = 'https://mirrors.aliyun.com/debian-security',
    [string]$Goproxy  = 'https://goproxy.cn,https://mirrors.aliyun.com/goproxy,direct',
    # libmpv source: github (fast via proxy) or sourceforge (what CI uses).
    [ValidateSet('github', 'sourceforge')]
    [string]$MpvSource = 'github',
    # Prefix prepended to the GitHub asset url. '' downloads from GitHub directly.
    [string]$MpvProxy  = 'https://gh-proxy.com/',
    # Skip resolution and use this .7z url (still gets the proxy prefix).
    [string]$MpvUrl    = '',
    # Skip `go vet` + `go test` (fast iteration; the cross-build itself still runs).
    [switch]$SkipTests,
    [switch]$Clean
)

$ErrorActionPreference = 'Stop'
$repo = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
if (-not (Test-Path (Join-Path $repo 'go.mod'))) { throw "not the mvplayer repo: $repo" }

function Step($msg) { Write-Host "`n==> $msg" -ForegroundColor Cyan }

Step "docker daemon"
$server = docker version --format '{{.Server.Version}}' 2>$null
if (-not $server) { throw 'docker daemon not reachable — start Docker Desktop first' }
Write-Host "server $server  hostOS=$([System.Environment]::OSVersion.VersionString)"

Step "image $Image"
docker image inspect $Image *> $null
if ($LASTEXITCODE -ne 0) {
    Write-Host "not local, pulling..."
    docker pull $Image
    if ($LASTEXITCODE -ne 0) { throw "cannot pull $Image" }
}

if ($Clean) {
    Step 'clean build/ + third_party/mpv'
    foreach ($p in 'build', 'third_party/mpv') {
        $full = Join-Path $repo $p
        if (Test-Path $full) {
            $resolved = (Resolve-Path $full).Path
            if (-not $resolved.StartsWith($repo, [StringComparison]::OrdinalIgnoreCase)) { throw "refusing to delete $resolved" }
            Remove-Item -Recurse -Force $resolved
            Write-Host "removed $p"
        }
    }
}

# The container cannot reach api.github.com directly here, so resolve the asset
# url on the host; the container only downloads the .7z (through -MpvProxy).
$mpvCached = Test-Path (Join-Path $repo 'third_party\mpv\libmpv-2.dll')
if (-not $MpvUrl -and -not $mpvCached -and $MpvSource -eq 'github') {
    Step 'resolve latest mpv-dev-x86_64 (host)'
    $rel = Invoke-RestMethod -Uri 'https://api.github.com/repos/shinchiro/mpv-winbuild-cmake/releases/latest' `
        -Headers @{ 'User-Agent' = 'mvplayer-build' } -TimeoutSec 30
    $asset = $rel.assets | Where-Object { $_.name -match '^mpv-dev-x86_64-\d+-git-[0-9a-f]+\.7z$' } |
        Select-Object -First 1
    if (-not $asset) { throw 'no mpv-dev-x86_64 asset in the latest shinchiro release' }
    $MpvUrl = $asset.browser_download_url
    Write-Host ("  {0}  {1} MiB" -f $asset.name, [math]::Round($asset.size / 1MB, 1))
}
if ($MpvUrl) { Write-Host "  url: $MpvUrl" -ForegroundColor DarkGray }
elseif ($mpvCached) { Write-Host '  libmpv already cached in third_party/mpv' -ForegroundColor DarkGray }

# The container copies exe + dll back into build\/src, so a running instance of the
# app (or anything else holding those files) makes the copy fail with EIO.
$stray = Get-Process -Name mvplayer -ErrorAction SilentlyContinue
if ($stray) {
    Step 'stop running mvplayer (it holds build\mvplayer.exe / libmpv-2.dll)'
    foreach ($proc in $stray) {
        Write-Host ("  pid {0}  {1}" -f $proc.Id, $proc.Path) -ForegroundColor DarkGray
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    }
    Start-Sleep -Milliseconds 800
}

# Named volumes keep the Go module cache and build cache across runs (~minutes saved).
Step "build in container"
$dockerArgs = @(
    'run', '--rm',
    '-v', "${repo}:/src",
    '-w', '/src',
    '-v', 'mvplayer-gomod:/go/pkg/mod',
    '-v', 'mvplayer-gobuild:/root/.cache/go-build',
    '-e', "APT_MIRROR=$AptMirror",
    '-e', "APT_MIRROR_SEC=$AptMirrorSec",
    '-e', "WANT_INSTALLER=$(if ($Installer) { 1 } else { 0 })",
    '-e', "SKIP_TESTS=$(if ($SkipTests) { 1 } else { 0 })",
    '-e', "GOPROXY=$Goproxy",
    '-e', 'GOSUMDB=off',
    '-e', 'GOTOOLCHAIN=local',
    '-e', "MPV_SOURCE=$MpvSource",
    '-e', "MPV_PROXY=$MpvProxy",
    '-e', "MPV_URL=$MpvUrl",
    $Image,
    'bash', '/src/scripts/docker/build-windows.sh'
)
Write-Host ("docker " + ($dockerArgs -join ' ')) -ForegroundColor DarkGray
& docker @dockerArgs
if ($LASTEXITCODE -ne 0) { throw "build failed (exit $LASTEXITCODE)" }

Step "artefacts"
Get-ChildItem (Join-Path $repo 'build') -File |
    Select-Object Name, @{n='MiB'; e={[math]::Round($_.Length / 1MB, 1)}}, LastWriteTime |
    Format-Table -AutoSize

$exe = Join-Path $repo 'build\mvplayer.exe'
if (Test-Path $exe) {
    $sig = [System.IO.File]::ReadAllBytes($exe)[0..1] -join ','
    Write-Host ("mvplayer.exe PE magic bytes: {0} (77,90 == MZ)" -f $sig)
}
