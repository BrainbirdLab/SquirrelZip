# SquirrelZip File compressor/archiver CLI-tool

Simple CLI-tool for compressing files.

## Usage
Build
```bash
./build
```

Run
```txt
./sq -i <file1,file2> -o <outputDir>
```

```txt
  -a    Read all files in the provided directory
  -d    File path to decompress (Only one file at once)
  -c    File paths to compress (Space separated file paths)
  -o string
        Output directory for compressed files (Optional)
  -p string
        Password for encryption (Optional)
  -v    Print version
```
## Examples
### Compress
#### Compress without password:
```txt
./sq -c file.txt file2.txt
```
#### Compress with password:
```txt
./sq -c file.txt file2.txt -p mySecurepass1234
```
#### Or compress the whole directory:
```txt
./sq -a folder
```
#### To provide output path use `-o` flag:
```txt
./sq -c file.txt -o output/files
```
### Decompress without password
```txt
./sq -d compressed.sq
```
### Decompress with password
```txt
./sq -d compressed.sq -p mySecurepass1234
```




