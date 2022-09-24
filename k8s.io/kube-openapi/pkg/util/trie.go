package util

type Trie struct {
	children map[byte]*Trie
	wordTail bool
	word     string
}

func NewTrie(list []string) Trie {
	ret := Trie{
		children: make(map[byte]*Trie),
		wordTail: false,
	}
	for _, v := range list {
		ret.Add(v)
	}
	return ret
}

func (t *Trie) Add(v string) {
	root := t
	for _, b := range []byte(v) {
		child, exists := root.children[b]
		if !exists {
			child = &Trie{
				children: make(map[byte]*Trie),
				wordTail: false,
			}
			root.children[b] = child
		}
		root = child
	}
	root.wordTail = true
	root.word = v
}

func (t *Trie) HasPrefix(v string) bool {
	_, has := t.GetPrefix(v)
	return has
}

func (t *Trie) GetPrefix(v string) (string, bool) {
	root := t
	if root.wordTail {
		return root.word, true
	}
	for _, b := range []byte(v) {
		child, exists := root.children[b]
		if !exists {
			return "", false
		}
		if child.wordTail {
			return child.word, true
		}
		root = child
	}
	return "", false
}
