@echo off
REM Build script to compile the Go program into an executable
echo Building resources.syso...
windres -o resources.syso resources.rc
echo Building chippi.exe...
go build -o chippi.exe
echo Build complete.