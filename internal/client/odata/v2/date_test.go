package v2

import (
	"testing"
	"time"
)

func TestParseDateLiteral_DateTimeOffset(t *testing.T) {
	want := time.Date(2019, 10, 4, 0, 0, 0, 0, time.UTC)
	for _, in := range []string{"/Date(1570147200000)/", "/Date(1570147200000+0000)/", "/Date(1570147200000+0120)/", "/Date(1570147200000-0060)/"} {
		got, err := ParseDateLiteral(in)
		if err != nil || !got.Equal(want) {
			t.Errorf("ParseDateLiteral(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, in := range []string{"/Date(1570147200000+ab)/", "2019-10-04T00:00:00Z"} {
		if _, err := ParseDateLiteral(in); err == nil {
			t.Errorf("ParseDateLiteral(%q) returned no error", in)
		}
	}
}
