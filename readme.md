# Chipmunk File compressor/archiver CLI-tool

Simple CLI-tool for compressing files using Huffman coding algorithm.

## Usage
Build
```bash
./build
```

Run
```txt
./chippi -i <file1,file2> -o <outputDir>
```

```txt
  -a    Read all files in the provided directory
  -d    File paths to decompress
  -c    File paths to compress
  -o string
        Output directory for compressed files (Optional)
  -p string
        Password for encryption (Optional)
  -v    Print version
```
## Examples
### Compress
#### Compress without password:
```bash
./chippi -c file.txt,file2.txt
```
#### Compress with password:
```bash
./chippi -c file.txt,file2.txt -p mySecurepass1234
```
#### Or compress the whole directory:
```bash
./chippi -a folder
```
#### To provide output path use `-o` flag:
```bash
./chippi -c file.txt -o output/files
```
### Decompress without password
```bash
./chippi -d compressed.arc
```
### Decompress with password
```bash
./chippi -d compressed.arc -p mySecurepass1234
```




