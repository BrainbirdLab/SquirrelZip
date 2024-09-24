package constants

const (

	BUFFER_SIZE = 256
	NO_PASSWORD byte = 43
	PASSWORD    byte = 57

	FILE_CREATE_ERROR = "failed to create file: %v\n"
	FILE_WRITE_ERROR = "failed to write file: %v\n"
	FILE_READ_ERROR = "failed to read file: %v\n"
	FILE_REMOVE_ERROR = "failed to remove file: %v\n"

	FILE_STAT_ERROR = "failed to get file info: %v\n"

	FILE_OPEN_ERROR = "failed to open file: %v\n"
	FILE_CLOSE_ERROR = "failed to close file: %v\n"

	ERROR_CREATE_DIR = "failed to create directory: %v\n"

	BUFFER_READ_ERROR = "failed to read buffer: %v\n"
	BUFFER_WRITE_ERROR = "failed to write buffer: %v\n"

	ERROR_DECOMPRESS = "failed to decompress file: %v\n"
	ERROR_COMPRESS = "failed to compress file: %v\n"

	FAILED_TO_ENCRYPT = "failed to encrypt file: %v\n"
	FAILED_TO_DECRYPT = "failed to decrypt file: %v\n"

	FAILED_GET_FREQ_MAP = "failed to get frequency map: %v\n"
	FAILED_BUILD_HUFFMAN_CODES = "failed to build huffman codes: %v\n"
	FAILED_READ_HUFFMAN_CODES = "failed to read huffman codes: %v\n"
	FAILED_WRITE_HUFFMAN_CODES = "failed to write huffman codes: %v\n"
)