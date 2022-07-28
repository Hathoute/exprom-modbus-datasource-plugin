package parser_test

import (
	"github.com/grafana/grafana-starter-datasource-backend/pkg/plugin/parser"
	"math"
	"testing"
)

func TestParser(t *testing.T) {

	type SubTestCase struct {
		format        string
		order         string
		expectedValue float64
	}

	type TestCase struct {
		rawBytes []byte
		values   []SubTestCase
	}

	cases := []TestCase{
		{
			rawBytes: []byte{0x42, 0x18, 0xE4, 0x00},
			values: []SubTestCase{
				{
					format:        "float32",
					order:         "ABCD",
					expectedValue: 38.2226563,
				},
				{
					format:        "float32",
					order:         "DCBA",
					expectedValue: 2.09471952e-38,
				},
				{
					format:        "float32",
					order:         "BADC",
					expectedValue: 2.5074362e-24,
				},
				{
					format:        "float32",
					order:         "CDAB",
					expectedValue: -9463783192163067625472,
				},
				{
					format:        "uint32",
					order:         "ABCD",
					expectedValue: 1108927488,
				},
				{
					format:        "uint32",
					order:         "DCBA",
					expectedValue: 14948418,
				},
				{
					format:        "uint32",
					order:         "BADC",
					expectedValue: 406978788,
				},
				{
					format:        "uint32",
					order:         "CDAB",
					expectedValue: 3825222168,
				},
			},
		},
	}

	isValid := func(val float64, expected float64) bool {
		const tolerance = 1e-7 // 5%
		return math.Abs(val-expected) < tolerance
	}

	for i, tc := range cases {
		bytes := tc.rawBytes
		for j, stc := range tc.values {
			parse := parser.GetBytesToDoubleParser(stc.format, stc.order)
			cp := make([]byte, len(bytes))
			copy(cp, bytes)
			val := parse(cp)
			if !isValid(val, stc.expectedValue) {
				t.Errorf("Value mismatch for test %d-%d: Expected %f, got %f", i, j, stc.expectedValue, val)
			}
		}
	}
}
