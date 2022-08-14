package parser

import (
	"encoding/binary"
	"errors"
	"math"
)

// Deprecated: The plugin does not need to parse data, as it is already saved as a double in the database.
func GetBytesToDoubleParser(format string, order string) func([]byte) (float64, error) {
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

	var parser binary.ByteOrder
	if isBigEndian {
		parser = binary.BigEndian
	} else {
		parser = binary.LittleEndian
	}

	return func(bytes []byte) (float64, error) {
		if len(order) != len(bytes) {
			return 0, errors.New("incompatible input bytes")
		}

		if isFlipped {
			flip(bytes)
		}

		switch format {
		case "int16":
			return float64(int16(parser.Uint16(bytes))), nil
		case "uint16":
			return float64(parser.Uint16(bytes)), nil
		case "int32":
			return float64(int32(parser.Uint32(bytes))), nil
		case "uint32":
			return float64(parser.Uint32(bytes)), nil
		case "float32":
			return float64(math.Float32frombits(parser.Uint32(bytes))), nil
		case "float64":
			return math.Float64frombits(parser.Uint64(bytes)), nil
		default:
			panic(errors.New("unknown format " + format))
		}
	}
}
