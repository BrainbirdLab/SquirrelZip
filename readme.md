# File compressor CLI-tool

Simple CLI-tool for compressing files using Huffman coding algorithm.

## Usage
Build
```bash
go build -o compress main.go
```

Run
```bash
./compress -i <file1,file2> -o <outputDir>
```


>  -a    Read all files in the test directory
>  -d    Decompress mode
>  -i string
>       Input files to be compressed
>  -o string
>       Output directory for compressed files (Optional)