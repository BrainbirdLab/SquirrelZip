@echo off
REM Cleanup script to delete 'test_output' and 'test_unzip' folders if they exist

REM Check if 'test_output' folder exists
if exist "test_files/test_output" (
    echo Deleting 'test_files/test_output' folder...
    rmdir /s /q "test_files/test_output"
)

REM Check if 'test_unzip' folder exists
if exist "test_files/test_decompressed_out" (
    echo Deleting 'test_files/test_decompressed_out' folder...
    rmdir /s /q "test_files/test_decompressed_out"
)

echo Cleanup complete.