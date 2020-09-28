package imdom

import (
	"encoding/binary"

	"github.com/vacwm/go-rapi"
	"google.golang.org/protobuf/proto"
)

// MessageLength converts the input to BigEndian format in 4 bytes.
func MessageLength(val uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, val)
	return b
}

// DecodeBytes assumes the first four bytes determine the length of the
// byte message.
func DecodeBytes(b []byte) (templateID int, err error) {
	messageType := &rti.MessageType{}
	if err := proto.Unmarshal(b[4:], messageType); err != nil {
		return 0, err
	}
	return int(*messageType.TemplateId), nil
}

// EncodeByteLength wraps bytes with a header containing
// the size in BigEndian format.
func EncodeByteLength(b []byte) []byte {
	message := make([]byte, 0, 4+len(b))
	message = append(message, MessageLength(uint32(len(b)))...)
	message = append(message, b...)
	return message
}
