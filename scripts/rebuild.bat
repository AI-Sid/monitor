@echo off

set TARGET_NAME=proxyMon.exe
set TARGET_DIR=bin
set GOOS=windows
set GOARCH=amd64
set SOURCE=cmd\app\main.go
set LDFLAGS=-H windowsgui

if exist "%TARGET_DIR%" (
    del /s /f /q "%TARGET_DIR%"
) else (
    mkdir "%TARGET_DIR%"
    if %errorlevel% neq 0 (
        echo Error on creating target directory
        exit /b 1
    )
)

go build -ldflags "%LDFLAGS%" -o "%TARGET_DIR%\%TARGET_NAME%" "%SOURCE%"
if %errorlevel% neq 0 (
    echo Error on build target file
    exit /b 1
)

echo Building completed
