@echo off
REM Test script to run all tests in the project
echo Running tests...
go test ./...
echo Tests complete.
REM clean up
./clean.bat