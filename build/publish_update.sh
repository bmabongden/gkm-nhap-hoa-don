#!/usr/bin/env bash
# Tao thu muc update/ (latest.json + GKM_NhapHoaDon.exe) tu app/GKM_NhapHoaDon.exe de dua len GitHub.
# App tren may nguoi dung doc update/latest.json de biet co ban moi, tai exe, kiem tra SHA-256 roi moi cai.
# Cach dung: bash build/publish_update.sh "ghi chu ban nay"   (chay sau build/build.sh), roi commit + push thu muc update/.
set -euo pipefail
cd "$(dirname "$0")/.."
EXE=app/GKM_NhapHoaDon.exe
[ -f "$EXE" ] || { echo "Chua co $EXE, hay chay build/build.sh truoc"; exit 1; }
VER=$(grep -oE 'const version = "[0-9]+\.[0-9]+\.[0-9]+"' source/cmd/gkm/main.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
mkdir -p update
cp "$EXE" update/GKM_NhapHoaDon.exe
SHA=$(sha256sum update/GKM_NhapHoaDon.exe | cut -d' ' -f1)
SIZE=$(stat -c %s update/GKM_NhapHoaDon.exe)
NOTES=${1:-}
python3 - "$VER" "$SHA" "$SIZE" "$NOTES" <<'PY'
import json,sys
v,sha,size,notes=sys.argv[1:5]
json.dump({"version":v,"sha256":sha,"size":int(size),"notes":notes},open("update/latest.json","w",encoding="utf-8"),ensure_ascii=False,indent=2)
open("update/latest.json","a").write("\n")
PY
echo "Da tao update/latest.json (ban $VER, sha256 $SHA, $SIZE byte)"
