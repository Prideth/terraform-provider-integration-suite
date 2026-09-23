package v2

import "testing"

func TestEncodeUTF8Hex(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"ascii word", "my alias", "6d7920616c696173"},
		{"documented SAP example", "smtp.mail.yahoo.com", "736d74702e6d61696c2e7961686f6f2e636f6d"},
		{"documented import example", "mycertificate", "6d796365727469666963617465"},
		{"documented rename example", "mykeypair", "6d796b657970616972"},
		{"documented ssh example", "baltimore cybertrust root", "62616c74696d6f7265206379626572747275737420726f6f74"},
		{"documented partner directory example", "agency1", "6167656e637931"},
		{"empty string", "", ""},
		{"space", " ", "20"},
		{"punctuation", "O'Reilly", "4f275265696c6c79"},
		{"semicolon", "Company A; Company B", "436f6d70616e7920413b20436f6d70616e792042"},
		{"slash", "Customer/Root", "437573746f6d65722f526f6f74"},
		{"backslash", "back\\slash", "6261636b5c736c617368"},
		{"unicode umlaut", "ümlaut", "c3bc6d6c617574"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := EncodeUTF8Hex(c.input)
			if got != c.want {
				t.Errorf("EncodeUTF8Hex(%q) = %q, want %q", c.input, got, c.want)
			}
		})
	}
}

func TestDecodeUTF8Hex_RoundTrip(t *testing.T) {
	inputs := []string{
		"my alias",
		"smtp.mail.yahoo.com",
		"",
		" ",
		"O'Reilly",
		"Company A; Company B",
		"Customer/Root",
		"back\\slash",
		"ümlaut",
		"日本語",
	}

	for _, in := range inputs {
		encoded := EncodeUTF8Hex(in)
		decoded, err := DecodeUTF8Hex(encoded)
		if err != nil {
			t.Errorf("DecodeUTF8Hex(EncodeUTF8Hex(%q)) error: %v", in, err)
			continue
		}
		if decoded != in {
			t.Errorf("round trip for %q produced %q", in, decoded)
		}
	}
}

func TestDecodeUTF8Hex_RejectsInvalidHex(t *testing.T) {
	if _, err := DecodeUTF8Hex("not-hex!"); err == nil {
		t.Error("DecodeUTF8Hex(\"not-hex!\") error = nil, want an error")
	}
	if _, err := DecodeUTF8Hex("abc"); err == nil {
		t.Error("DecodeUTF8Hex(\"abc\") (odd length) error = nil, want an error")
	}
}

func TestDecodeUTF8Hex_AcceptsUppercase(t *testing.T) {
	got, err := DecodeUTF8Hex("736D74702E6D61696C2E7961686F6F2E636F6D")
	if err != nil {
		t.Fatalf("DecodeUTF8Hex() error: %v", err)
	}
	if got != "smtp.mail.yahoo.com" {
		t.Errorf("DecodeUTF8Hex() = %q, want smtp.mail.yahoo.com", got)
	}
}
