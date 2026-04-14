package server


const DATA_BLOCK_SIZE = 4
const MAX_DATA_SIZE = 4294967296

const ADDITIONAL_BLOCK_YES = 0xCC
const ADDITIONAL_BLOCK_NO = 0xAA

const (
	ERR      byte = iota + 1
	USR_CONN
	TOK_CONN
	ECHO_S
	ECHO_R
	FLE_LST
)

