# SquirrelZip

SquirrelZip is a powerful command-line file compression and encryption tool written in Go. It provides secure file compression with optional password protection.

## Features

- File compression with multiple algorithm support
- Password-based encryption for compressed files
- Cross-platform support (Windows, Linux, macOS)
- Command-line interface
- Progress tracking and compression ratio reporting
- Secure file handling with automatic cleanup

# Supported Algorithms
- Huffman
- LZMA

## Installation

### Pre-built Binaries

You can download pre-built binaries for your platform from the [Releases](https://github.com/itsfuad/SquirrelZip/releases) page.

### Building from Source

1. Ensure you have Go 1.22 or later installed
2. Clone the repository:
   ```bash
   git clone https://github.com/itsfuad/SquirrelZip.git
   cd SquirrelZip
   ```
3. Build the project:
   ```bash
   go build
   ```

## Usage

### Compressing Files

```bash
SquirrelZip -c [files...] -o [output_dir] -p [password] -a [algorithm] -all
```

Options:
- `-c`: Input files or directory to be compressed
- `-o`: Output directory (optional)
- `-p`: Password for encryption (optional)
- `-a`: Compression algorithm (optional, defaults to "huffman")
- `-all`: Read all files in the input directory (optional)

### Decompressing Files

```bash
SquirrelZip -d [file] -o [output_dir] -p [password]
```

Options:
- `-d`: Input file to decompress
- `-o`: Output directory (optional)
- `-p`: Password for decryption (required if file was encrypted)

### Additional Options
- `-v`: Print version
- `-h`: Print help

## Development

### Project Structure

```
SquirrelZip/
├── compressor/     # Compression algorithms
├── encryption/     # Encryption/decryption logic
├── utils/         # Utility functions
├── constants/     # Constants and messages
├── main.go        # Main application entry
└── test_files/    # Test files
```

### Building

```bash
# Windows
build.bat

# Linux/macOS
go build
```

### Testing

```bash
# Windows
test.bat
testRun.bat
testRunNopass.bat

# Linux/macOS
go test ./...
```

## License

This project is licensed under the terms specified in the LICENSE file.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- Built with Go
- Uses standard library compression algorithms
- Implements secure encryption for file protection
