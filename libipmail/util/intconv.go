package util

import (
	"fmt"
)

func BytesToInt32(bytes []byte) (int32, error) {
	if len(bytes) > 4 {
		return 0, fmt.Errorf("int32 can only have a max of 4 "+
			"bytes. bytes passed in: %d", len(bytes))
	}
	var result int32 = 0
	for _, val := range bytes {
		result = (result << 8) | int32(val)
	}
	return result, nil
}

func BytesToInt64(bytes []byte) (int64, error) {
	if len(bytes) > 8 {
		return 0, fmt.Errorf("int64 can only have a max of 8 "+
			"bytes. bytes passed in: %d", len(bytes))
	}
	var result int64 = 0
	for _, val := range bytes {
		result = (result << 8) | int64(val)
	}
	return result, nil
}

func BytesToUint32(bytes []byte) (uint32, error) {
	if len(bytes) > 4 {
		return 0, fmt.Errorf("uint32 can only have a max of 4 "+
			"bytes. bytes passed in: %d", len(bytes))
	}
	var result uint32 = 0
	for _, val := range bytes {
		result = (result << 8) | uint32(val)
	}
	return result, nil
}

func BytesToUint64(bytes []byte) (uint64, error) {
	if len(bytes) > 8 {
		return 0, fmt.Errorf("uint64 can only have a max of 8 "+
			"bytes. bytes passed in: %d", len(bytes))
	}
	var result uint64 = 0
	for _, val := range bytes {
		result = (result << 8) | uint64(val)
	}
	return result, nil
}

func intToBytes(int int64, size uint8) []byte {
	result := make([]byte, size)
	for i := int16(size - 1); i >= 0; i -= 1 {
		result[i] = byte(int & 0xff)
		int >>= 8
	}
	return result
}

func Int32ToBytes(int int32) []byte {
	return intToBytes(int64(int), 4)
}

func Int64ToBytes(int int64) []byte {
	return intToBytes(int, 8)
}

func Uint32ToBytes(int uint32) []byte {
	return intToBytes(int64(int), 4)
}

func Uint64ToBytes(int uint64) []byte {
	return intToBytes(int64(int), 8)
}
