package rulecompiler

import (
	"encoding/json"
	"reflect"
	"testing"
)

type fakeFormat struct{}

func (fakeFormat) Compile(json.RawMessage) (Result, error) { return Result{}, nil }

func TestCompilerUsesRegisteredFormatAdapters(t *testing.T) {
	compiler, err := New(map[string]Adapter{"kazumi": Kazumi{}, "future": fakeFormat{}})
	if err != nil {
		t.Fatal(err)
	}
	if formats := compiler.Formats(); !reflect.DeepEqual(formats, []string{"future", "kazumi"}) {
		t.Fatalf("formats = %#v", formats)
	}
	if _, err := compiler.Compile("missing", json.RawMessage(`{}`)); err == nil {
		t.Fatal("unregistered format unexpectedly compiled")
	}
}

func TestKazumiCompilesCommonRuleAndKeepsCompatibilityAdviceNonBlocking(t *testing.T) {
	compiler, err := New(map[string]Adapter{"kazumi": Kazumi{}})
	if err != nil {
		t.Fatal(err)
	}
	result, err := compiler.Compile(" KAZUMI ", json.RawMessage(`{
		"api": 3,
		"name": "Example Anime",
		"baseURL": "https://anime.example",
		"searchURL": "/search?keyword=@keyword",
		"searchList": "//ul/li",
		"searchName": "/a/text()",
		"searchResult": "/a/@href",
		"chapterRoads": "//div[@class='roads']",
		"chapterResult": "/a/@href",
		"muliSources": true,
		"useWebview": true,
		"useNativePlayer": false,
		"referer": "https://anime.example/"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Supported() {
		t.Fatalf("ordinary compatibility warnings unexpectedly blocked rule: %#v", result.Diagnostics)
	}
	if result.Plan.Format != "kazumi" || result.Plan.FormatLevel != 3 || !result.Plan.Title.MultipleRoads {
		t.Fatalf("plan = %#v", result.Plan)
	}
	if len(result.Diagnostics) != 2 || result.Diagnostics[0].Severity == SeverityUnsupported || result.Diagnostics[1].Severity == SeverityUnsupported {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestKazumiOnlyBlocksCapabilitiesThatBreakProductContracts(t *testing.T) {
	result, err := (Kazumi{}).Compile(json.RawMessage(`{
		"api": 5,
		"name": "Scripted",
		"baseURL": "https://anime.example",
		"searchURL": "/search/@keyword",
		"searchList": "//li",
		"searchName": "/a/text()",
		"searchResult": "/a/@href",
		"chapterRoads": "//main",
		"chapterResult": "//a/@href",
		"antiCrawlerConfig": {"script": "window.challenge()"},
		"adBlocker": ["ad.ts"]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if result.Supported() {
		t.Fatalf("contract-breaking rule unexpectedly supported: %#v", result.Diagnostics)
	}
	if len(result.Diagnostics) != 2 || result.Diagnostics[0].Severity != SeverityUnsupported || result.Diagnostics[1].Severity != SeverityUnsupported {
		t.Fatalf("diagnostics = %#v", result.Diagnostics)
	}
}

func TestKazumiRejectsStructurallyInvalidRule(t *testing.T) {
	_, err := (Kazumi{}).Compile(json.RawMessage(`{"api":1,"name":"Broken","baseURL":"http://127.0.0.1"}`))
	if err == nil {
		t.Fatal("invalid rule unexpectedly compiled")
	}
}
