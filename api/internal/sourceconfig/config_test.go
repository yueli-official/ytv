package sourceconfig

import "testing"

func TestLoadAddsFixtureOnlyByExplicitOptIn(t *testing.T) {
	if _, err := Load("", false); err == nil {
		t.Fatal("empty production source definitions unexpectedly accepted")
	}
	definitions, err := Load("", true)
	if err != nil || len(definitions) != 1 || definitions[0].Key != "fixture" {
		t.Fatalf("definitions/error = %#v / %v", definitions, err)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	_, err := Load(`[{"key":"x","name":"X","kind":"json_catalog","endpoint":"https://example.test","unknown":true}]`, false)
	if err == nil {
		t.Fatal("unknown deployment field unexpectedly accepted")
	}
}

func TestLoadDefaultsHomeVisibilityAndAcceptsExplicitOptOut(t *testing.T) {
	definitions, err := Load(`[
		{"key":"a","name":"A","kind":"json_catalog","endpoint":"https://a.example.test","enabled":true,"priority":1},
		{"key":"b","name":"B","kind":"json_catalog","endpoint":"https://b.example.test","installed":false,"enabled":true,"homeVisible":false,"priority":2}
	]`, false)
	if err != nil || !definitions[0].HomeVisible || definitions[0].InitiallyUninstalled || definitions[1].HomeVisible || !definitions[1].InitiallyUninstalled {
		t.Fatalf("definitions/error = %#v / %v", definitions, err)
	}
}

func TestLoadCompilesDeclarativeKazumiDefinition(t *testing.T) {
	definitions, err := Load(`[{
		"key":"anime","name":"Anime","kind":"declarative","enabled":true,
		"playbackHosts":["media.example"],"ruleFormat":"kazumi",
		"rule":{
			"api":1,"name":"Anime","baseURL":"https://anime.example","searchURL":"/search/@keyword",
			"searchList":"//li","searchName":"/a/text()","searchResult":"/a/@href",
			"chapterRoads":"//main","chapterResult":"/a","useNativePlayer":true,
			"ytv":{"mediaXPath":"//video/source/@src"}
		}
	}]`, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(definitions) != 1 || definitions[0].Endpoint != "https://anime.example" || definitions[0].Rule == nil || definitions[0].Rule.Playback.MediaXPath == "" {
		t.Fatalf("definitions = %#v", definitions)
	}
}

func TestLoadKeepsUnsupportedCapabilitiesAsOperatorNotices(t *testing.T) {
	definitions, err := Load(`[{
		"key":"scripted","name":"Scripted","kind":"declarative","enabled":true,"ruleFormat":"kazumi",
		"rule":{
			"api":5,"name":"Scripted","baseURL":"https://anime.example","searchURL":"/search/@keyword",
			"searchList":"//li","searchName":"/a/text()","searchResult":"/a/@href",
			"chapterRoads":"//main","chapterResult":"/a","antiCrawlerConfig":{"script":"run()"}
		}
	}]`, false)
	if err != nil || len(definitions) != 1 || len(definitions[0].Compatibility) != 1 || definitions[0].Compatibility[0].Level != "unsupported" {
		t.Fatalf("definitions/error = %#v / %v", definitions, err)
	}
}
