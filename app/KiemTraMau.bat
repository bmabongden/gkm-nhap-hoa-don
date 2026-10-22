@echo off
rem Kiem tra nhanh: chay 3 mau (Mau 1/2/3) qua chinh GKM_NhapHoaDon.exe va so tung o voi ket qua mau.
chcp 65001 >nul
cd /d "%~dp0"
if not exist "%~dp0..\samples\Mau 1.xlsx" (
  echo Khong co thu muc samples (hoa don mau khong dua len kho cong khai). Chep thu muc samples tu ban day du vao day de kiem tra mau.
  pause
  exit /b 2
)
if not exist "..\logs" mkdir "..\logs"
"%~dp0GKM_NhapHoaDon.exe" --selftest --samples "%~dp0..\samples" > "%~dp0..\logs\kiemtra_mau.txt" 2>&1
set RC=%ERRORLEVEL%
type "%~dp0..\logs\kiemtra_mau.txt"
echo.
if "%RC%"=="0" (echo KET QUA: DAT) else (echo KET QUA: CO LOI - xem logs\kiemtra_mau.txt)
pause
exit /b %RC%
