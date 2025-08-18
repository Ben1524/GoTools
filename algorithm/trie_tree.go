package algorithm

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool // 是否是一个完整的单词
}

type TrieTree struct {
	root *TrieNode
}

func NewTrieTree() *TrieTree {
	return &TrieTree{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
			isEnd:    false,
		},
	}
}

func (t *TrieTree) Insert(word string) {
	node := t.root
	for _, char := range word { // 遍历每个字符
		if _, exists := node.children[char]; !exists { // 如果当前字符不存在，则创建一个新的 TrieNode
			node.children[char] = &TrieNode{
				children: make(map[rune]*TrieNode),
				isEnd:    false,
			}
		}
		node = node.children[char]
	}
	node.isEnd = true
}

func (t *TrieTree) Search(word string) bool {
	node := t.root
	for _, char := range word { // 遍历每个字符
		if _, exists := node.children[char]; !exists { // 如果当前字符不存在，则返回 false
			return false
		}
		node = node.children[char]
	}
	return node.isEnd // 返回是否是一个完整的单词
}

func (t *TrieTree) StartsWith(prefix string) bool {
	node := t.root
	for _, char := range prefix { // 遍历每个字符
		if _, exists := node.children[char]; !exists { // 如果当前字符不存在，则返回 false
			return false
		}
		node = node.children[char]
	}
	return true // 返回是否存在以 prefix 为前缀的单词
}

func (t *TrieTree) Delete(word string) {
	t.deleteHelper(t.root, word, 0)
}

func (t *TrieTree) deleteHelper(node *TrieNode, word string, index int) bool {
	if index == len(word) { // 到达单词的末尾
		if !node.isEnd { // 如果当前节点不是单词的终止符，说明单词不存在
			return false // 单词不存在
		}
		node.isEnd = false
		return len(node.children) == 0 // 如果没有子节点，返回 true 以删除该节点
	}

	char := rune(word[index])
	childNode, exists := node.children[char]
	if !exists {
		return false // 单词不存在
	}

	shouldDeleteChild := t.deleteHelper(childNode, word, index+1) // 递归删除子节点保证单词存在
	if shouldDeleteChild {
		delete(node.children, char)                   // 删除子节点
		return len(node.children) == 0 && !node.isEnd // 如果没有子节点且不是终止符，返回 true 以删除该节点
	}
	return false
}
