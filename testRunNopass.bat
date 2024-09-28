@echo off
echo Running compression command
go run main.go -all -c "test_files/input" -o "test_files/test_output"
REM wait for the command to finish
echo Compression command finished
echo Running decompression command
go run main.go -d "test_files/test_output/example.sq" -o "test_files/test_decompressed_out"
echo Decompression command finished
./clean