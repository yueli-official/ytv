package browsercapture

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestLiveCapture(t *testing.T) {
	if os.Getenv("YTV_TEST_LIVE_BROWSER_CAPTURE") != "true" {
		t.Skip("set YTV_TEST_LIVE_BROWSER_CAPTURE=true for the live browser capture smoke")
	}
	resolver, err := New(Options{Executable: os.Getenv("YTV_BROWSER_EXECUTABLE"), MaxConcurrent: 1, Timeout: 30 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer resolver.Close()
	cases := []string{
		"https://www.moonci.com/anime/228/play/2-1.html",
	}
	for _, target := range cases {
		t.Run(target, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
			defer cancel()
			result, err := resolver.Resolve(ctx, Request{PageURL: target})
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			t.Logf("media=%s", result.URL)
		})
	}
}
