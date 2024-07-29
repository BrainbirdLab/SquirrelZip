@echo off
REM Build script to compile the Go program into an executable
echo Building resources.syso...
windres -o resources.syso resources.rc
echo Building sq.exe...
go build -o sq.exe
echo Build complete.