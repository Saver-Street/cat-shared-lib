package types

import (
	"sort"
	"testing"
)

func TestTriePutGet(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()

	if isNew := tr.Put("foo", 1); !isNew {
		t.Error("Put(foo) should be new")
	}
	if isNew := tr.Put("bar", 2); !isNew {
		t.Error("Put(bar) should be new")
	}
	if isNew := tr.Put("foo", 3); isNew {
		t.Error("Put(foo) again should not be new")
	}

	v, ok := tr.Get("foo")
	if !ok || v != 3 {
		t.Errorf("Get(foo) = (%d, %v); want (3, true)", v, ok)
	}
	v, ok = tr.Get("bar")
	if !ok || v != 2 {
		t.Errorf("Get(bar) = (%d, %v); want (2, true)", v, ok)
	}
	_, ok = tr.Get("baz")
	if ok {
		t.Error("Get(baz) should return false")
	}
}

func TestTrieDelete(t *testing.T) {
	t.Parallel()
	tr := NewTrie[string]()
	tr.Put("hello", "world")
	tr.Put("help", "me")

	if !tr.Delete("hello") {
		t.Error("Delete(hello) should return true")
	}
	if tr.Has("hello") {
		t.Error("Has(hello) after delete should be false")
	}
	if tr.Len() != 1 {
		t.Errorf("Len() = %d; want 1", tr.Len())
	}

	// help should still exist
	v, ok := tr.Get("help")
	if !ok || v != "me" {
		t.Errorf("Get(help) = (%q, %v); want (me, true)", v, ok)
	}

	// delete non-existent
	if tr.Delete("nothing") {
		t.Error("Delete(nothing) should return false")
	}
	// delete prefix that is not a key
	if tr.Delete("hel") {
		t.Error("Delete(hel) should return false (not a key)")
	}
}

func TestTrieHas(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	tr.Put("abc", 1)

	if !tr.Has("abc") {
		t.Error("Has(abc) = false; want true")
	}
	if tr.Has("ab") {
		t.Error("Has(ab) = true; want false")
	}
	if tr.Has("abcd") {
		t.Error("Has(abcd) = true; want false")
	}
}

func TestTrieHasPrefix(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	tr.Put("foobar", 1)
	tr.Put("foobaz", 2)

	if !tr.HasPrefix("foo") {
		t.Error("HasPrefix(foo) = false; want true")
	}
	if !tr.HasPrefix("foobar") {
		t.Error("HasPrefix(foobar) = false; want true")
	}
	if tr.HasPrefix("baz") {
		t.Error("HasPrefix(baz) = true; want false")
	}
}

func TestTrieWithPrefix(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	tr.Put("app", 1)
	tr.Put("apple", 2)
	tr.Put("apply", 3)
	tr.Put("banana", 4)

	entries := tr.WithPrefix("app")
	sort.Slice(entries, func(i, j int) bool { return entries[i].Key < entries[j].Key })

	if len(entries) != 3 {
		t.Fatalf("WithPrefix(app) len = %d; want 3", len(entries))
	}
	want := []struct {
		key string
		val int
	}{{"app", 1}, {"apple", 2}, {"apply", 3}}
	for i, e := range entries {
		if e.Key != want[i].key || e.Value != want[i].val {
			t.Errorf("[%d] = (%q, %d); want (%q, %d)", i, e.Key, e.Value, want[i].key, want[i].val)
		}
	}
}

func TestTrieWithPrefixEmpty(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	tr.Put("abc", 1)
	entries := tr.WithPrefix("xyz")
	if len(entries) != 0 {
		t.Errorf("WithPrefix(xyz) = %v; want empty", entries)
	}
}

func TestTrieEmptyKey(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	tr.Put("", 42)

	v, ok := tr.Get("")
	if !ok || v != 42 {
		t.Errorf("Get('') = (%d, %v); want (42, true)", v, ok)
	}
	if !tr.Has("") {
		t.Error("Has('') = false; want true")
	}
}

func TestTrieLen(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	if tr.Len() != 0 {
		t.Errorf("Len() = %d; want 0", tr.Len())
	}
	tr.Put("a", 1)
	tr.Put("b", 2)
	if tr.Len() != 2 {
		t.Errorf("Len() = %d; want 2", tr.Len())
	}
	tr.Delete("a")
	if tr.Len() != 1 {
		t.Errorf("Len() = %d; want 1", tr.Len())
	}
}

func TestTrieDeleteCleansNodes(t *testing.T) {
	t.Parallel()
	tr := NewTrie[int]()
	tr.Put("abcdef", 1)
	tr.Delete("abcdef")
	// After deleting the only key, internal nodes should be cleaned up
	if tr.HasPrefix("a") {
		t.Error("HasPrefix(a) after full delete should be false")
	}
}

func BenchmarkTriePut(b *testing.B) {
	keys := []string{"foo", "bar", "baz", "foobar", "foobaz", "barfoo"}
	for range b.N {
		tr := NewTrie[int]()
		for i, k := range keys {
			tr.Put(k, i)
		}
	}
}

func BenchmarkTrieGet(b *testing.B) {
	tr := NewTrie[int]()
	keys := []string{"foo", "bar", "baz", "foobar", "foobaz", "barfoo"}
	for i, k := range keys {
		tr.Put(k, i)
	}
	b.ResetTimer()
	for i := range b.N {
		tr.Get(keys[i%len(keys)])
	}
}

func FuzzTriePutGet(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("abc")
	f.Fuzz(func(t *testing.T, key string) {
		tr := NewTrie[int]()
		tr.Put(key, 42)
		v, ok := tr.Get(key)
		if !ok || v != 42 {
			t.Errorf("Get(%q) = (%d, %v); want (42, true)", key, v, ok)
		}
		tr.Delete(key)
		if tr.Has(key) {
			t.Errorf("Has(%q) after delete = true", key)
		}
	})
}
