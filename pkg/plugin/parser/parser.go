package parser

import (
	"encoding/binary"
	"errors"
	"math"
)

func GetBytesToDoubleParser(format string, order string) func([]byte) float64 {
	flip := func(bytes []byte) {
		// TODO: Any drawbacks in mutating this?
		i := 1
		for i < len(bytes) {
			bytes[i], bytes[i-1] = bytes[i-1], bytes[i]
			i += 2
		}
	}

	isBigEndian := false
	isFlipped := false

	switch order {
	case "AB", "ABCD", "BADC", "ABCDEFGH", "BADCFEHG":
		isBigEndian = true
	}

	switch order {
	case "BADC", "CDAB", "BADCFEHG", "GHEFCDAB":
		isFlipped = true
	}

	return func(bytes []byte) float64 {
		var parser binary.ByteOrder
		if isBigEndian {
			parser = binary.BigEndian
		} else {
			parser = binary.LittleEndian
		}

		if isFlipped {
			flip(bytes)
		}

		switch format {
		case "int16":
			return float64(int16(parser.Uint16(bytes)))
		case "uint16":
			return float64(parser.Uint16(bytes))
		case "int32":
			return float64(int32(parser.Uint32(bytes)))
		case "uint32":
			return float64(parser.Uint32(bytes))
		case "float32":
			return float64(math.Float32frombits(parser.Uint32(bytes)))
		case "float64":
			return math.Float64frombits(parser.Uint64(bytes))
		default:
			panic(errors.New("unknown format " + format))
		}
	}
}
