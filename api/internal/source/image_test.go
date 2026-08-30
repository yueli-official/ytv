package source

import "testing"

func TestAllowedImageURLRequiresDeploymentHostAndRejectsPrivateAddresses(t *testing.T) {
	definition := Definition{ImageHosts: []string{"images.example", "*.cdn.example"}}
	if got := AllowedImageURL(definition, "https://images.example/poster.jpg"); got == "" {
		t.Fatal("allowlisted poster was rejected")
	}
	for _, value := range []string{
		"http://images.example/poster.jpg", "https://unlisted.example/poster.jpg", "https://127.0.0.1/poster.jpg",
	} {
		if got := AllowedImageURL(definition, value); got != "" {
			t.Fatalf("unsafe poster %q allowed as %q", value, got)
		}
	}
}
