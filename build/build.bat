@echo off
rem GKM - Nhap Hoa Don: build ban portable .exe tren Windows. Khong can Internet (thu vien da nam trong source\vendor).
rem Yeu cau: da cai Go 1.23 tro len (https://go.dev/dl/). Toan bo cache build dat trong build\ cua du an, khong dung o C.
setlocal
cd /d "%~dp0.."
where go >nul 2>nul
if errorlevel 1 (
  echo Chua cai Go. Cai Go 1.23+ roi chay lai file nay.
  pause
  exit /b 1
)
set "GKM_ROOT=%CD%"
set "GOCACHE=%GKM_ROOT%\build\gocache"
set "GOPATH=%GKM_ROOT%\build\gopath"
set "GOTMPDIR=%GKM_ROOT%\build\tmp"
set "GOFLAGS=-mod=vendor"
set "GOTOOLCHAIN=local"
set "CGO_ENABLED=0"
set "GOOS=windows"
set "GOARCH=amd64"
if not exist "%GOTMPDIR%" mkdir "%GOTMPDIR%"
copy /y "templates\MauFileNhapHang.xlsx" "source\cmd\gkm\template.xlsx" >nul
pushd source
echo [1/3] Chay kiem thu loi chuyen doi (3 mau) ...
go test ./internal/... ./cmd/...
if errorlevel 1 (
  popd
  echo Kiem thu that bai, khong build.
  pause
  exit /b 1
)
echo [2/3] Build GKM_NhapHoaDon.exe ...
set "LDF=-s -w -H=windowsgui"
if defined GKM_REPO set "LDF=%LDF% -X gkm/internal/update.DefaultRepo=%GKM_REPO%"
go build -trimpath -ldflags "%LDF%" -o "..\app\GKM_NhapHoaDon.exe" .\cmd\gkm
if errorlevel 1 (
  popd
  echo Build that bai.
  pause
  exit /b 1
)
popd
echo [3/3] Chay kiem tra 3 mau bang chinh file .exe vua build ...
call "app\KiemTraMau.bat"
echo.
echo Xong. File: %GKM_ROOT%\app\GKM_NhapHoaDon.exe
pause
