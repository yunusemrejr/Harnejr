package events

import "testing"

func TestRedactStringRemovesCommonSecrets(t *testing.T) {
	input := "Authorization: Bearer sk-secret123456789 and api_key=nvapi-secret123456789"
	redacted := RedactString(input)
	if redacted == input || contains(redacted, "sk-secret") || contains(redacted, "nvapi-secret") {
		t.Fatalf("secret was not redacted: %s", redacted)
	}
}

func contains(s string, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
