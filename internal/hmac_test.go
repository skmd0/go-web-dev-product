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
	type args struct {
		arg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default hmac key", args{"3NTSaJPS84ll2MqUyY5gUc1e85802GqXAPrYFp9ZvYw="}, "MIzRSOPTyTDFpnL38OP8iP8JPdwpPOawvZlPcdEdKHI="},
	}

	for _, tt := range tests {
		key := "my-hmac-secret-key"
		h := NewHMAC(key)
		t.Run(tt.name, func(t *testing.T) {
			got := h.Hash(tt.args.arg)
			if got != tt.want {
				t.Errorf("hmac.Hash(%q) = %q; want %q", tt.args.arg, got, tt.want)
			}
		})
	}
}
