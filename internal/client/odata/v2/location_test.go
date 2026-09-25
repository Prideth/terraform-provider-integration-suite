package v2

import "testing"

func TestServiceRoot(t *testing.T) {
	cases := map[string]string{
		"":             "https://tenant.example/api/v1",
		"myedge":       "https://tenant.example/location/myedge/api/v1",
		"plant-a_2.eu": "https://tenant.example/location/plant-a_2.eu/api/v1",
	}
	for loc, want := range cases {
		got, err := ServiceRoot("https://tenant.example/", loc)
		if err != nil || got != want {
			t.Errorf("ServiceRoot(%q) = %q, %v; want %q", loc, got, err, want)
		}
	}
	for _, bad := range []string{"../api", "a/b", "a b", "-lead", "a?x=1", "a%2F"} {
		if _, err := ServiceRoot("https://tenant.example", bad); err == nil {
			t.Errorf("ServiceRoot(%q) returned no error", bad)
		}
	}
}
