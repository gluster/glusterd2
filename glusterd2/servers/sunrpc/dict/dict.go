package dict

import (
	"bytes"
	"encoding/binary"
	"errors"
)

/*
   Gluster dict serialized format:
   ------------------------------------------------------
   |  count | key len | val len | key     \0 | value    |
   ------------------------------------------------------
       4        4         4       <key len>   <value len>

NOTE: The "key len" computed does not include the NULL character at the end.
      All values are serialized to strings and "val len" takes the NULL
      character into consideration.
*/

const (
	dictHeaderLen = 4
)

// Unserialize unmarshals a slice of bytes into a map[string]string
// Consumers of the map should typecast/extract information from the
// map values which are of string type
func Unserialize(buf []byte) (map[string]string, error) {

	newDict := make(map[string]string)
	tmpHeader := make([]byte, dictHeaderLen)

	var keyLen uint32
	var valueLen uint32

	reader := bytes.NewReader(buf)

	// Extract dict count
	reader.Read(tmpHeader)
	count := int(binary.BigEndian.Uint32(tmpHeader))

	if count < 0 {
		return nil, errors.New("invalid dict count")
	}

	for i := 0; i < count; i++ {
		// Read key length
		reader.Read(tmpHeader)
		keyLen = binary.BigEndian.Uint32(tmpHeader)

		// Read value length
		reader.Read(tmpHeader)
		valueLen = binary.BigEndian.Uint32(tmpHeader)

		// Read key
		key := make([]byte, keyLen+1) // +1 for '/0'
		reader.Read(key)

		// Read value
		value := make([]byte, valueLen)
		reader.Read(value)

		// Strings aren't NULL terminated in Go
		newDict[string(key[:len(key)-1])] = string(value[:len(value)-1])
	}

	return newDict, nil
}

// Serialize marshals a map[string]string into a slice of bytes.
func Serialize(dict map[string]string) ([]byte, error) {

	dictSerializedSize, err := getSerializedDictLen(dict)
	if err != nil {
		return nil, err
	}

	// Force buffer to have fixed size by setting desired capacity
	// but a length of 0
	buffer := bytes.NewBuffer(make([]byte, 0, dictSerializedSize))
	tmpHeader := make([]byte, dictHeaderLen)
	var totalBytesWritten int

	// Write dict count
	count := len(dict)
	binary.BigEndian.PutUint32(tmpHeader, uint32(count))
	bytesWritten, err := buffer.Write(tmpHeader)
	if err != nil {
		return nil, err
	}
	totalBytesWritten += bytesWritten

	for key, value := range dict {

		// write key length
		binary.BigEndian.PutUint32(tmpHeader, uint32(len(key)))
		bytesWritten, err := buffer.Write(tmpHeader)
		if err != nil {
			return nil, err
		}
		totalBytesWritten += bytesWritten

		// write value length
		binary.BigEndian.PutUint32(tmpHeader, uint32(len(value)+1))
		bytesWritten, err = buffer.Write(tmpHeader)
		if err != nil {
			return nil, err
		}
		totalBytesWritten += bytesWritten

		// write key + '\0'
		bytesWritten, err = buffer.Write([]byte(key))
		if err != nil {
			return nil, err
		}
		totalBytesWritten += bytesWritten
		bytesWritten, err = buffer.Write([]byte("\x00"))
		if err != nil {
			return nil, err
		}
		totalBytesWritten += bytesWritten

		// write value + '\0'
		// Values are serialized to strings and strings in C are NULL terminated.
		bytesWritten, err = buffer.Write([]byte(value))
		if err != nil {
			return nil, err
		}
		totalBytesWritten += bytesWritten
		bytesWritten, err = buffer.Write([]byte("\x00"))
		if err != nil {
			return nil, err
		}
		totalBytesWritten += bytesWritten
	}

	if dictSerializedSize != totalBytesWritten {
		return nil, errors.New("dict serialized size mismatch")
	}

	return buffer.Bytes(), nil
}

func getSerializedDictLen(dict map[string]string) (int, error) {

	if dict == nil || len(dict) == 0 {
		return 0, errors.New("nil or empty dict")
	}

	totalSize := int(dictHeaderLen) // dict count
	for key, value := range dict {
		// Key length and value length
		totalSize += dictHeaderLen + dictHeaderLen
		totalSize += (len(key) + 1) + (len(value) + 1)
	}

	return totalSize, nil
}
