@echo off
set ARG=%1

if "%ARG%" == "" (
    set ARG=all
)

python scripts/doBuild.py proxyMon %ARG% "-X 'main.Build=true'"
