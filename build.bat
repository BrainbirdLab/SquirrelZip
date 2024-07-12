@echo off
REM Build script to compile the Go program into an executable
echo Building resources.syso...
windres -o resources.syso resources.rc
echo Building piarch.exe...
go build -o piarch.exe
echo Build complete.