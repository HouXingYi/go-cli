@echo off
setlocal enabledelayedexpansion

set DIST_DIR=dist

if exist "%DIST_DIR%" rmdir /s /q "%DIST_DIR%"
mkdir "%DIST_DIR%"

call :build_variant zn-ent ./cmd/zn-ent
call :build_variant zn-eco ./cmd/zn-eco

echo Build artifacts are in %DIST_DIR%/
exit /b 0

:build_variant
set APP_NAME=%1
set ENTRY=%2

call :build linux amd64
call :build linux arm64
call :build darwin amd64
call :build darwin arm64
call :build windows amd64 .exe
exit /b 0

:build
set GOOS=%1
set GOARCH=%2
set EXT=%3
set CGO_ENABLED=0
set OUTPUT=%DIST_DIR%\%APP_NAME%-%GOOS%-%GOARCH%%EXT%

echo Building %OUTPUT%
go build -o "%OUTPUT%" "%ENTRY%"
if errorlevel 1 exit /b 1
exit /b 0
