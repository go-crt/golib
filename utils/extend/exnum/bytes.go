package exnum

import (
	"encoding/binary"
	"math"
)

//all are bigendian format

func BytesToUint16(b []byte) uint16 {
	return binary.BigEndian.Uint16(b)
}

func Uint16ToBytes(u uint16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, u)
	return buf
}

func BytesToUint32(b []byte) uint32 {
	return binary.BigEndian.Uint32(b)
}

func Uint32ToBytes(u uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, u)
	return buf
}

func BytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

func Uint64ToBytes(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}

func BytesToInt16(b []byte) int16 {
	return int16(binary.BigEndian.Uint16(b))
}

func Int16ToBytes(u int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(u))
	return buf
}

func BytesToInt32(b []byte) int32 {
	return int32(binary.BigEndian.Uint32(b))
}

func Int32ToBytes(u int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(u))
	return buf
}

func BytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func Int64ToBytes(u int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(u))
	return buf
}

func BytesToFloat32(b []byte) float32 {
	bits := binary.LittleEndian.Uint32(b)
	return math.Float32frombits(bits)
}

func Float32ToBytes(u float32) []byte {
	bits := math.Float32bits(u)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, bits)
	return buf
}

func BytesToFloat64(b []byte) float64 {
	bits := binary.LittleEndian.Uint64(b)
	return math.Float64frombits(bits)
}

func Float64ToBytes(u float64) []byte {
	bits := math.Float64bits(u)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, bits)
	return buf
}
