package v2

import (
	"errors"
	"testing"
)

type testPackage struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
}

func TestDecodeEntity(t *testing.T) {
	body := []byte(`{"d": {"Id": "UTILITIES", "Name": "Utilities"}}`)

	var got testPackage
	if err := DecodeEntity(body, &got); err != nil {
		t.Fatalf("DecodeEntity() error: %v", err)
	}

	want := testPackage{ID: "UTILITIES", Name: "Utilities"}
	if got != want {
		t.Errorf("DecodeEntity() = %+v, want %+v", got, want)
	}
}

func TestDecodeCollection(t *testing.T) {
	body := []byte(`{"d": {"results": [{"Id": "A", "Name": "A"}, {"Id": "B", "Name": "B"}]}}`)

	var got []testPackage
	if err := DecodeCollection(body, &got); err != nil {
		t.Fatalf("DecodeCollection() error: %v", err)
	}

	if len(got) != 2 || got[0].ID != "A" || got[1].ID != "B" {
		t.Errorf("DecodeCollection() = %+v", got)
	}
}

func TestDecodeEntity_InvalidJSON(t *testing.T) {
	var got testPackage
	if err := DecodeEntity([]byte("not json"), &got); err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

func TestDecodeEntity_EmptyBody(t *testing.T) {
	var v struct{ Name string }
	for _, body := range [][]byte{nil, []byte(""), []byte("  \n")} {
		if err := DecodeEntity(body, &v); !errors.Is(err, ErrEmptyBody) {
			t.Errorf("DecodeEntity(%q) error = %v, want ErrEmptyBody", body, err)
		}
	}
}
