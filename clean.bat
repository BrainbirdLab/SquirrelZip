@echo off
REM Cleanup script to delete 'test_output' and 'test_unzip' folders if they exist

REM Check if 'test_output' folder exists
if exist "test_zip" (
    echo Deleting 'test_output' folder...
    rmdir /s /q "test_zip"
)

REM Check if 'test_unzip' folder exists
if exist "test_unzip" (
    echo Deleting 'test_unzip' folder...
    rmdir /s /q "test_unzip"
)

echo Cleanup complete.