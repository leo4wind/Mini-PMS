@echo off
cd /d "%~dp0"
mise exec go -- go run ./cmd/server
