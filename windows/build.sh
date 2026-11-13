#!/usr/bin/env bash
# Gera dist/ObservatorioLimeira.exe e dist/ObservatorioLimeira-Instalador.exe
# Requisitos: Go 1.22+, makensis (NSIS 3)
set -e
cd "$(dirname "$0")"
go get github.com/jchv/go-webview2@latest golang.org/x/sys@latest
go mod tidy
go run github.com/akavel/rsrc@latest -manifest app.manifest -ico app.ico -arch amd64 -o rsrc_windows_amd64.syso
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-H windowsgui -s -w" -o dist/ObservatorioLimeira.exe .
makensis -INPUTCHARSET UTF8 installer.nsi
ls -la dist
