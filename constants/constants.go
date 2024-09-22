package constants

const (

	BUFFER_SIZE = 256
	NO_PASSWORD byte = 43
	PASSWORD    byte = 57

	FILE_CREATE_ERROR = "failed to create file: %v"
	FILE_WRITE_ERROR = "failed to write file: %v"
	FILE_READ_ERROR = "failed to read file: %v"
	FILE_REMOVE_ERROR = "failed to remove file: %v"

	FILE_STAT_ERROR = "failed to get file info: %v"

	FILE_OPEN_ERROR = "failed to open file: %v"
	FILE_CLOSE_ERROR = "failed to close file: %v"

	ERROR_CREATE_DIR = "failed to create directory: %v"

	BUFFER_READ_ERROR = "failed to read buffer: %v"
	BUFFER_WRITE_ERROR = "failed to write buffer: %v"

	ERROR_DECOMPRESS = "failed to decompress file: %v"
	ERROR_COMPRESS = "failed to compress file: %v"

	FAILED_GET_FREQ_MAP = "failed to get frequency map: %v"
	FAILED_BUILD_HUFFMAN_CODES = "failed to build huffman codes: %v"
	FAILED_READ_HUFFMAN_CODES = "failed to read huffman codes: %v"
	FAILED_WRITE_HUFFMAN_CODES = "failed to write huffman codes: %v"
)