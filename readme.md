# SquirrelZip File Compressor/Archiver CLI Tool

Simple CLI tool for compressing and decompressing files.

## Usage

### Build

./build

### Run

./sq -c <file1,file2> -o <outputDir>

  -v      Print version information
  -c      Input files or directory to be compressed [strings] (Space separated)
  -o      Output directory for compressed/decompressed files (Optional)
  -a      Algorithm to use for compression (Optional) [string]
  -p      Password for encryption (Optional) [string]
  -all    Read all files in the provided directory (Optional)
  -d      Input file to decompress [strings] (Space separated)
  -h      Print help

## Examples

### Compress
#### Compress without password:
./sq -c file.txt file2.txt

#### Compress with password:
./sq -c file.txt file2.txt -p mySecurepass1234

#### Or compress the whole directory:
./sq -all folder

#### To provide an output path use the `-o` flag:
./sq -c file.txt -o output/files

### Decompress without password:
./sq -d compressed.sq

### Decompress with password:
./sq -d compressed.sq -p mySecurepass1234
