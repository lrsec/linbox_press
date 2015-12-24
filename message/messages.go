package message

import (
	"encoding/binary"
	"encoding/json"
)

func BuildBytes(rrType RequestResponseType, content interface{}) ([]byte, error) {
	contentByte, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	var length uint32 = uint32(len(contentByte))

	result := make([]byte, length+2+4)

	binary.BigEndian.PutUint16(result[0:2], uint16(rrType))
	binary.BigEndian.PutUint32(result[2:6], length)
	copy(result[6:], contentByte)

	return result, nil
}
