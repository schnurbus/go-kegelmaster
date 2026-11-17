package apitoken

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	a, err := generate()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := generate()
	if a == b {
		t.Fatal("tokens must be unique")
	}
	if !strings.HasPrefix(a, Prefix) || len(a) != len(Prefix)+43 {
		t.Fatalf("unexpected token format %q", a)
	}
	if strings.ContainsAny(a, "+/=") {
		t.Fatalf("token must be URL-safe: %q", a)
	}
}

func TestHashToken(t *testing.T) {
	if HashToken("kgm_a") == HashToken("kgm_b") || len(HashToken("kgm_a")) != 64 {
		t.Fatal("unexpected hash")
	}
}
