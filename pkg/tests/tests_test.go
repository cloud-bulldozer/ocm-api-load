package tests

import (
	"encoding/json"
	"testing"
)

func TestCertificateBody(t *testing.T) {
	buff := certificatesBody()
	type payload struct {
		Type string `json:"type"`
		Arch string `json:"arch"`
	}
	var pl payload
	err := json.Unmarshal(buff, &pl)
	if err != nil {
		t.Fatalf("error unmarshaling payload %v", pl)
	}
	if pl.Type == "" && pl.Arch == "" {
		t.Fatalf("certificates body did not return payload to be unmarshaled %v", pl)
	}
}
