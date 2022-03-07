package communication

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializerBinaryMessage(t *testing.T) {
	header := map[string]string{
		"id": "123",
	}
	msgBytes := SerializeBinaryMessage(header, []byte("lorem ipsum"))
	fmt.Printf("t: %v\n", msgBytes)

	assert.Equal(t, []byte{49, 50, 42, 123, 34, 105, 100, 34, 58, 34, 49, 50, 51, 34, 125, 108, 111, 114, 101, 109, 32, 105, 112, 115, 117, 109}, msgBytes)
}
func TestDeserializerBinaryMessage(t *testing.T) {
	header := map[string]string{
		"id": "123",
	}
	msgBytes := SerializeBinaryMessage(header, []byte("lorem ipsum"))
	deserializedHeaders, deserializedBytes := DeserializeBinaryMessage(msgBytes)

	assert.Len(t, deserializedHeaders, 1)
	assert.Equal(t, deserializedHeaders["id"], "123")
	assert.Equal(t, []byte{108, 111, 114, 101, 109, 32, 105, 112, 115, 117, 109}, deserializedBytes)
}
