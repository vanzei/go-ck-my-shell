package core

type TrieNode struct {
	Children map[rune]*TrieNode
	IsWord   bool
}

func NewTrieNode() *TrieNode {
	return &TrieNode{Children: make(map[rune]*TrieNode)}
}

type Trie struct {
	Root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{Root: NewTrieNode()}
}

func (t *Trie) Insert(word string) {
	node := t.Root
	for _, ch := range word {
		if node.Children[ch] == nil {
			node.Children[ch] = NewTrieNode()
		}
		node = node.Children[ch]
	}
	node.IsWord = true
}

func (t *Trie) FindWordsWithPrefix(prefix string) []string {
	node := t.Root
	for _, ch := range prefix {
		if node.Children[ch] == nil {
			return nil
		}
		node = node.Children[ch]
	}
	var results []string
	var dfs func(*TrieNode, []rune)
	dfs = func(n *TrieNode, path []rune) {
		if n.IsWord {
			results = append(results, prefix+string(path))
		}
		for ch, child := range n.Children {
			dfs(child, append(path, ch))
		}
	}
	dfs(node, []rune{})
	return results
}

func (t *Trie) IsWord(word string) bool {
	node := t.Root
	for _, ch := range word {
		if node.Children[ch] == nil {
			return false
		}
		node = node.Children[ch]
	}
	return node.IsWord
}
