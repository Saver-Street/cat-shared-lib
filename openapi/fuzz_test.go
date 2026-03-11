package openapi

import (
	"encoding/json"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func FuzzNewSpec(f *testing.F) {
	f.Add("Pet Store API", "1.0.0")
	f.Add("", "")
	f.Add("unicode-タイトル", "v2.0.0-beta")
	f.Add("title with \"quotes\"", "1.0")
	f.Add("title\nwith\nnewlines", "v1")
	f.Fuzz(func(t *testing.T, title, version string) {
		s := NewSpec(title, version)
		if s == nil {
			t.Fatal("NewSpec returned nil")
		}
		// JSON must not panic and must produce valid JSON.
		data, err := s.JSON()
		testkit.RequireNoError(t, err)
		if !json.Valid(data) {
			t.Error("JSON() produced invalid JSON")
		}
	})
}

func FuzzSpecFluent(f *testing.F) {
	f.Add("API", "1.0", "A description", "https://api.example.com", "Production")
	f.Add("", "", "", "", "")
	f.Add("日本語", "v2", "説明", "http://localhost", "ローカル")
	f.Fuzz(func(t *testing.T, title, version, desc, url, serverDesc string) {
		s := NewSpec(title, version).
			WithDescription(desc).
			AddServer(url, serverDesc)
		if s == nil {
			t.Fatal("fluent chain returned nil")
		}
		data, err := s.JSON()
		testkit.RequireNoError(t, err)
		if !json.Valid(data) {
			t.Error("JSON() produced invalid JSON")
		}
	})
}

func FuzzOperation(f *testing.F) {
	f.Add("List pets", "Returns all pets", "listPets", "pets")
	f.Add("", "", "", "")
	f.Add("summary \"quoted\"", "desc\nwith\nnewlines", "op-id", "tag1")
	f.Fuzz(func(t *testing.T, summary, desc, opID, tag string) {
		op := NewOperation(summary).
			WithDescription(desc).
			WithOperationID(opID).
			WithTags(tag)
		if op == nil {
			t.Fatal("operation chain returned nil")
		}

		// Add to a spec and serialize.
		s := NewSpec("test", "1.0").
			AddPath("/test", "get", op)
		data, err := s.JSON()
		testkit.RequireNoError(t, err)
		if !json.Valid(data) {
			t.Error("JSON() produced invalid JSON")
		}
	})
}
