package constants

// Packet constants.
const DATA_BLOCK_SIZE = 4
const MAX_DATA_SIZE = 4294967296

const CONTINUATION_BYTE_YES = 0xCC
const CONTINUATION_BYTE_NO = 0xAA

const (
	ERR      byte = iota
	USR_CONN
	TOK_CONN
	CLOSE
	TOK_RES
	OK
	ECHO_S
	ECHO_R
	FLE_LST
	FLE_DWN
	FLE_UP
)

// File constants.
const DATA_FOLDER = "data"
const CREDENTIALS_FILE = "credentials.sql"
const CONFIG_FILE = "config.yaml"
const LOG_FILE = "server_log.log"