#!/usr/bin/env bash
# Build ban .exe Windows tu Linux/macOS (cross-compile, khong can Windows). Can Go 1.23+.
set -euo pipefail
cd "$(dirname "$0")/.."
export GOFLAGS=-mod=vendor GOTOOLCHAIN=local CGO_ENABLED=0
cp templates/MauFileNhapHang.xlsx source/cmd/gkm/template.xlsx
# GKM_REPO (tuy chon, dang owner/repo) la kho GitHub mac dinh cua chuc nang Cap nhat online.
LDFLAGS="-s -w -H=windowsgui"
[ -n "${GKM_REPO:-}" ] && LDFLAGS="$LDFLAGS -X gkm/internal/update.DefaultRepo=$GKM_REPO"
( cd source && go test ./internal/... ./cmd/... && GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$LDFLAGS" -o ../app/GKM_NhapHoaDon.exe ./cmd/gkm )
file app/GKM_NhapHoaDon.exe 2>/dev/null || true
echo "Da build app/GKM_NhapHoaDon.exe"
