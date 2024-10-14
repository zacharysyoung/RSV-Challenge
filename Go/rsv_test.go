package main

import (
	"reflect"
	"testing"
)

// Transcribed from TestFiles/Valid_001.rsv
var (
	validRSV  = []byte("\x48\x65\x6c\x6c\x6f\xff\xf0\x9f\x8c\x8e\xff\xfe\xff\xff\xfd\x41\x00\x42\x0a\x43\xff\x54\x65\x73\x74\x20\xf0\x9d\x84\x9e\xff\xfd\xfd\xff\xfd")
	validRows = [][]NullableString{
		{Str("Hello"), Str("🌎"), Null(), Str("")},
		{Str("A\x00B\nC"), Str("Test 𝄞")},
		{},
		{Str("")},
	}
)

func TestEncode(t *testing.T) {
	b, err := EncodeRsv(validRows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(b, validRSV) {
		t.Errorf("DecodeRsv(validRows)\n  got %v\n want %v", b, validRSV)
	}
}

func TestDecode(t *testing.T) {
	rows, err := DecodeRsv(validRSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != len(validRows) {
		t.Errorf("len(rows)\n  got %d\n want %d", len(rows), len(validRows))
	}

	for i := 0; i < len(rows); i++ {
		if len(rows[i]) != len(validRows[i]) {
			t.Errorf("len(rows[%d])\n  got %d\n want %d", i, len(rows), len(validRows))
		}
		for j := 0; j < len(rows[i]); j++ {
			if !reflect.DeepEqual(rows[i][j], validRows[i][j]) {
				t.Errorf("rows[%d][%d]\n  got %v\n want %v", i, j, rows[i][j], validRows[i][j])
			}
		}
	}
}

func TestDecodeUsingSplit(t *testing.T) {
	rows, err := DecodeRsvUsingSplit(validRSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != len(validRows) {
		t.Errorf("len(rows)\n  got %d\n want %d", len(rows), len(validRows))
	}

	for i := 0; i < len(rows); i++ {
		if len(rows[i]) != len(validRows[i]) {
			t.Errorf("len(rows[%d])\n  got %d\n want %d", i, len(rows), len(validRows))
		}
		for j := 0; j < len(rows[i]); j++ {
			if !reflect.DeepEqual(rows[i][j], validRows[i][j]) {
				t.Errorf("rows[%d][%d]\n  got %v\n want %v", i, j, rows[i][j], validRows[i][j])
			}
		}
	}
}

func BenchmarkEncodeDecode(b *testing.B) {
	b.Run("Encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			EncodeRsv(validRows)
		}
	})
	b.Run("Decode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DecodeRsv(validRSV)
		}
	})
}

func TestEscapeJSON(t *testing.T) {
	testCases := []struct {
		s, want string
	}{
		{"Köln", `"Köln"`},                   // ASCII & non-ASCII, unmodified
		{"Foo\\nBar", `"Foo\\nBar"`},         // Control char escaped
		{"Rec1\x1fRec2", `"Rec1\u001fRec2"`}, // Non-standard control chars escaped, https://stackoverflow.com/a/18782271/246801
	}
	for _, tc := range testCases {
		if got := EscapeJsonString(tc.s); got != tc.want {
			t.Errorf("\nEscapeJsonString(%q)\n  got %v\n want %v", tc.s, got, tc.want)
		}
	}
}

func TestIncomplete(t *testing.T) {
	// omit terminal rowSep from:
	//   | COL1 | COL2 |
	//   | R1C1 | R1C2 |
	//   | R2C1 | R2C2 |
	//
	b := []byte("COL1\xffCOL2\xff\xfdR1C1\xffR1C2\xff\xfdR2C1\xffR2C2\xff")

	if _, got := DecodeRsv(b); got != errIncompleteDoc {
		t.Errorf("for invalid RSV:\n  got %v\n want %v", got, errIncompleteDoc)
	}
}

func TestIsValidRSV(t *testing.T) {
	testCases := []struct {
		s    []byte
		want bool
	}{
		{[]byte{}, true},
		{[]byte{rowTerm}, true},
		{[]byte{valTerm, rowTerm}, true},
		{[]byte{nullVal, valTerm, rowTerm}, true},

		{[]byte{'H', 'e', 'l', 'l', 'o', valTerm, rowTerm}, true},
		{[]byte{'H', 'e', 'l', 'l', 0xc3, 0xb6, valTerm, rowTerm}, true}, // Hellö
		{[]byte{0xf0, 0x9f, 0x8c, 0x8d, valTerm, rowTerm}, true},         // 🌍

		{[]byte{'H', 'e', 'l', 'l', 'o', valTerm}, false},
		{[]byte{'H', 'e', 'l', 'l', 'o', rowTerm}, false},
		{[]byte{'H', 'e', 'l', 'l', 'o'}, false},
	}
	for _, tc := range testCases {
		if got := IsValidRsv(tc.s); got != tc.want {
			t.Errorf("IsValidRsv(%q)=%t; want %t", tc.s, got, tc.want)
		}
	}
}
