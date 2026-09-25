package v2

import "testing"

func TestQuery_Encode(t *testing.T) {
	q := Query{
		Filter: "Name eq 'foo'",
		Select: []string{"Id", "Name"},
		Top:    10,
		Skip:   5,
	}
	got := q.Encode()
	want := "%24filter=Name+eq+%27foo%27&%24select=Id%2CName&%24skip=5&%24top=10"
	if got != want {
		t.Errorf("Encode() = %q, want %q", got, want)
	}
}

func TestEscapeLiteral(t *testing.T) {
	got := EscapeLiteral("O'Brien")
	want := "O''Brien"
	if got != want {
		t.Errorf("EscapeLiteral() = %q, want %q", got, want)
	}
}

func TestFilterEquals(t *testing.T) {
	got := FilterEquals("Name", "it's a test")
	want := "Name eq 'it''s a test'"
	if got != want {
		t.Errorf("FilterEquals() = %q, want %q", got, want)
	}
}

func TestCompositeKeyPredicate(t *testing.T) {
	got, err := CompositeKeyPredicate("Id", "metering", "Version", "1.0.1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "(Id='metering',Version='1.0.1')"
	if got != want {
		t.Errorf("CompositeKeyPredicate() = %q, want %q", got, want)
	}
}

func TestCompositeKeyPredicate_OddArgs(t *testing.T) {
	if _, err := CompositeKeyPredicate("Id"); err == nil {
		t.Fatal("expected an error for an odd number of arguments")
	}
}

func TestCompositeKeyPredicate_EscapesValues(t *testing.T) {
	got, err := CompositeKeyPredicate("Id", "O'Brien")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "(Id='O''Brien')"
	if got != want {
		t.Errorf("CompositeKeyPredicate() = %q, want %q", got, want)
	}
}

func TestBuildPath(t *testing.T) {
	got := BuildPath("IntegrationPackages", KeyPredicate("UTILITIES"), "")
	want := "IntegrationPackages('UTILITIES')"
	if got != want {
		t.Errorf("BuildPath() = %q, want %q", got, want)
	}

	got = BuildPath("IntegrationPackages", "", (Query{Top: 1}).Encode())
	want = "IntegrationPackages?%24top=1"
	if got != want {
		t.Errorf("BuildPath() = %q, want %q", got, want)
	}
}

func TestInt64KeyPredicate(t *testing.T) {
	cases := map[string]string{"1901": "(1901L)", "0": "(0L)", "-3": "(-3L)"}
	for in, want := range cases {
		got, err := Int64KeyPredicate(in)
		if err != nil || got != want {
			t.Errorf("Int64KeyPredicate(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "abc", "1'", "1)/Other(2", "99999999999999999999"} {
		if got, err := Int64KeyPredicate(in); err == nil {
			t.Errorf("Int64KeyPredicate(%q) = %q, want an error", in, got)
		}
	}
}
