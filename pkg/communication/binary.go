package communication

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type BinaryHeader map[string]string

// SerializeBinaryMessage into format: <headers_len>*<bytes_headers><bytes_data>
func SerializeBinaryMessage(headers BinaryHeader, data []byte) []byte {
	byteHeaders, _ := json.Marshal(headers)
	lenByteHeaders := len(byteHeaders)
	prefixMsg := []byte(fmt.Sprintf("%d*", lenByteHeaders))
	prefixMsg = append(prefixMsg, byteHeaders...)

	msg := make([]byte, 0)
	msg = append(prefixMsg, data...)
	return msg
}

// DeserializeBinaryMessage msg in format: <headers_len>*<bytes_headers><bytes_data>
func DeserializeBinaryMessage(msg []byte) (map[string]string, []byte) {
	// 1. headers_len part
	headerLenBytes := make([]byte, 0)
	separatorIdx := 0
	for i, b := range msg {
		if b == 42 { // asterisk
			separatorIdx = i
			break
		}
		headerLenBytes = append(headerLenBytes, b)
	}
	headerLen, _ := strconv.Atoi(string(headerLenBytes))

	// 2. fetch headers
	headers := make(map[string]string)
	if headerLen > 0 { // has headers
		headerBytes := msg[separatorIdx+1 : separatorIdx+1+headerLen]
		json.Unmarshal(headerBytes, &headers)
	}
	dataBytes := msg[separatorIdx+1+headerLen:]
	return headers, dataBytes
}
