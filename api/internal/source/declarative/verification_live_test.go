package declarative_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/yueli-official/ytv/api/internal/source"
	"github.com/yueli-official/ytv/api/internal/source/declarative"
	"github.com/yueli-official/ytv/api/internal/source/rulecompiler"
	"github.com/yueli-official/ytv/api/internal/source/rules"
	"github.com/yueli-official/ytv/api/internal/source/verification"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Live evidence is opt-in: upstream availability and challenges vary over time.
func TestLiveMutefunVerification(t *testing.T) {
	if os.Getenv("YTV_TEST_LIVE_VERIFICATION") != "true" {
		t.Skip("explicit upstream verification probe only")
	}
	raw, err := os.ReadFile("../rulecatalog/snapshot/kazumi.json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Rules map[string]json.RawMessage `json:"rules"`
	}
	if err = json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	compiled, err := (rulecompiler.Kazumi{}).Compile(snapshot.Rules["mutefun"])
	if err != nil {
		t.Fatal(err)
	}
	rules.ApplySpecialPlayback(&compiled.Plan)
	// Probe only: this does not add mutefun to the deployment-approved catalog.
	definition := source.Definition{Key: "kazumi-mutefun", Endpoint: compiled.Plan.BaseURL, Rule: &compiled.Plan}
	adapter, _ := declarative.New(&http.Client{Timeout: 20 * time.Second})
	_, err = adapter.Search(context.Background(), definition, source.Query{Text: "海贼王", Page: 1})
	var required *verification.Required
	if !errors.As(err, &required) || required.Kind != verification.Image {
		t.Fatalf("expected observed image challenge, got %v", err)
	}
}

// This probe keeps cookies only in memory. A human supplies the image answer in
// the task artifact directory; absence of an answer is not successful verification.
func TestLiveMutefunManualSession(t *testing.T) {
	directory := os.Getenv("YTV_TEST_LIVE_CAPTCHA_DIR")
	if directory == "" {
		t.Skip("requires explicit task directory and human answer")
	}
	if _, err := os.Stat(filepath.Join(directory, "answer.txt")); !os.IsNotExist(err) {
		t.Fatal("use a fresh task directory without a previous answer")
	}
	raw, err := os.ReadFile("../rulecatalog/snapshot/kazumi.json")
	if err != nil {
		t.Fatal(err)
	}
	var snapshot struct {
		Rules map[string]json.RawMessage `json:"rules"`
	}
	if err = json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	compiled, err := (rulecompiler.Kazumi{}).Compile(snapshot.Rules["mutefun"])
	if err != nil {
		t.Fatal(err)
	}
	rules.ApplySpecialPlayback(&compiled.Plan)
	definition := source.Definition{Key: "kazumi-mutefun", Endpoint: compiled.Plan.BaseURL, Rule: &compiled.Plan}
	flow, err := declarative.NewMutefunSearchVerification(http.DefaultClient, definition, source.Query{Text: "海贼王", Page: 1})
	if err != nil {
		t.Fatal(err)
	}
	sessions := verification.NewSessions()
	defer sessions.Close()
	challenge, err := sessions.Start(context.Background(), "manual-probe", flow)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(directory, "captcha.png"), challenge.Image.Bytes, 0600); err != nil {
		t.Fatal(err)
	}
	t.Log("captcha.png ready; waiting up to 120 seconds for human answer.txt")
	deadline := time.NewTimer(120 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline.C:
			t.Skip("manual answer absent; successful live verification remains unverified")
		case <-ticker.C:
			answer, readErr := os.ReadFile(filepath.Join(directory, "answer.txt"))
			if os.IsNotExist(readErr) {
				continue
			}
			if readErr != nil {
				t.Fatal(readErr)
			}
			result, submitErr := sessions.Submit(context.Background(), "manual-probe", challenge.ID, string(answer))
			if submitErr != nil {
				t.Fatal(submitErr)
			}
			if !result.Accepted {
				t.Fatal("human answer rejected")
			}
			t.Logf("human verification accepted; original search resumed with %d results", len(result.Page.Items))
			return
		}
	}
}
