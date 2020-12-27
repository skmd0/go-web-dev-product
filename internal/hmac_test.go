package internal

import (
	"testing"
)

func TestNewHMAC(t *testing.T) {
	got := NewHMAC("").hmac
	if got == nil {
		t.Fatal("NewHMAC returned a nil result instead of hash.Hash")
	}
}

func TestHMAC_Hash(t *testing.T) {
	key := "my-hmac-secret-key"
	h := NewHMAC(key)

	input := "3NTSaJPS84ll2MqUyY5gUc1e85802GqXAPrYFp9ZvYw="
	got := h.Hash(input)
	want := "MIzRSOPTyTDFpnL38OP8iP8JPdwpPOawvZlPcdEdKHI="
	if got != want {
		t.Errorf("hmac.Hash(%q) = %q; want %q", input, got, want)
	}

	inputEmpty := ""
	gotEmpty := h.Hash(inputEmpty)
	if gotEmpty == "" {
		t.Error("hmac.Hash() = ''; want non empty value")
	}
}
