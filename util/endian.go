//util - a package about sequence data
//Using LittleEndian

package util

import (
	"encoding/binary"
	"io"
)

//LittleEndian [01 02 03 04] => [04 03 02 01]
var DefaultEndian = binary.LittleEndian

//ReadUint8 - Read 1 byte
func ReadUint8(r io.Reader) (uint8, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return uint8(buf[0]), nil
}

//ReadUint16 - Read 2 byte
func ReadUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return DefaultEndian.Uint16(buf), nil
}

//ReadUint32 - Read 4 byte
func ReadUint32(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return DefaultEndian.Uint32(buf), nil
}

//ReadUint64 - Read 8 byte
func ReadUint64(r io.Reader) (uint64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	return DefaultEndian.Uint64(buf), nil
}

//ReadData - Following format |data len |data/payload|
func ReadData(r io.Reader) ([]byte, error) {
	len, err := ReadUint32(r)
	if err != nil {
		return nil, err
	}
	//read data with len
	buf := make([]byte, len)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

//WriteUint8 - Write 1 byte
func WriteUint8(w io.Writer, val uint8) error {
	buf := []byte{val}
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

//WriteUint16 - Write 2 byte
func WriteUint16(w io.Writer, val uint16) error {
	buf := make([]byte, 2)
	DefaultEndian.PutUint16(buf, val)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

//WriteUint32 - Write 4 byte
func WriteUint32(w io.Writer, val uint32) error {
	buf := make([]byte, 4)
	DefaultEndian.PutUint32(buf, val)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

//WriteUint64 - Write 8 byte
func WriteUint64(w io.Writer, val uint64) error {
	buf := make([]byte, 8)
	DefaultEndian.PutUint64(buf, val)
	if _, err := w.Write(buf); err != nil {
		return err
	}
	return nil
}

//WriteData - Following format |data len |data/payload|
func WriteData(w io.Writer, buf []byte) error {
	len := len(buf)

	if err := WriteUint32(w, uint32(len)); err != nil {
		return err
	}

	if _, err := w.Write(buf); err != nil {
		return err
	}

	return nil
}
