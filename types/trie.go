package types

// Trie is a prefix tree for string keys with values of type T.
type Trie[T any] struct {
	children map[byte]*Trie[T]
	value    *T
	size     int
}

// NewTrie creates an empty trie.
func NewTrie[T any]() *Trie[T] {
	return &Trie[T]{children: make(map[byte]*Trie[T])}
}

// Put associates key with value. It returns true if the key was new.
func (t *Trie[T]) Put(key string, val T) bool {
	node := t
	for i := range len(key) {
		c := key[i]
		child, ok := node.children[c]
		if !ok {
			child = &Trie[T]{children: make(map[byte]*Trie[T])}
			node.children[c] = child
		}
		node = child
	}
	isNew := node.value == nil
	node.value = &val
	if isNew {
		t.size++
	}
	return isNew
}

// Get returns the value for key and whether it was found.
func (t *Trie[T]) Get(key string) (val T, ok bool) {
	node := t.find(key)
	if node == nil || node.value == nil {
		var zero T
		return zero, false
	}
	return *node.value, true
}

// Delete removes a key. It returns true if the key existed.
func (t *Trie[T]) Delete(key string) bool {
	return t.delete(t, key, 0)
}

func (t *Trie[T]) delete(root *Trie[T], key string, depth int) bool {
	if depth == len(key) {
		if t.value == nil {
			return false
		}
		t.value = nil
		root.size--
		return true
	}
	c := key[depth]
	child, ok := t.children[c]
	if !ok {
		return false
	}
	found := child.delete(root, key, depth+1)
	if found && child.value == nil && len(child.children) == 0 {
		delete(t.children, c)
	}
	return found
}

// Has reports whether the key exists.
func (t *Trie[T]) Has(key string) bool {
	node := t.find(key)
	return node != nil && node.value != nil
}

// HasPrefix reports whether any key starts with prefix.
func (t *Trie[T]) HasPrefix(prefix string) bool {
	return t.find(prefix) != nil
}

// WithPrefix returns all key-value pairs whose key starts with prefix.
func (t *Trie[T]) WithPrefix(prefix string) []TrieEntry[T] {
	node := t.find(prefix)
	if node == nil {
		return nil
	}
	var entries []TrieEntry[T]
	node.collect(prefix, &entries)
	return entries
}

// TrieEntry holds a key-value pair returned by iteration methods.
type TrieEntry[T any] struct {
	Key   string
	Value T
}

func (t *Trie[T]) collect(prefix string, entries *[]TrieEntry[T]) {
	if t.value != nil {
		*entries = append(*entries, TrieEntry[T]{Key: prefix, Value: *t.value})
	}
	for c, child := range t.children {
		child.collect(prefix+string(c), entries)
	}
}

// Len returns the number of keys stored.
func (t *Trie[T]) Len() int { return t.size }

func (t *Trie[T]) find(key string) *Trie[T] {
	node := t
	for i := range len(key) {
		child, ok := node.children[key[i]]
		if !ok {
			return nil
		}
		node = child
	}
	return node
}
