package internal

import "testing"

func TestNBytes(t *testing.T) {
	token := "Dx0zh8ecHIzBzmXYAirRhmF8LKys1wzFRF-iqW1bEt8="
	got, err := NBytes(token)
	if err != nil {
		t.Fatal(err)
	}
	want := 32
	if got != want {
		t.Errorf("NBytes(%q) = %d; want %d", token, got, want)
	}
}
