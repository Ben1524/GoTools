package algorithm

import "testing"

// 测试基本插入和搜索功能
func TestTrie_BasicOperations(t *testing.T) {
	trie := NewTrieTree()

	// 测试初始状态
	if trie.Search("") {
		t.Error("新创建的Trie不应包含空字符串")
	}

	// 插入并验证单词
	trie.Insert("apple")
	if !trie.Search("apple") {
		t.Error("应能搜索到已插入的apple")
	}
	if trie.Search("app") {
		t.Error("不应搜索到未插入的app")
	}

	// 测试前缀匹配
	if !trie.StartsWith("app") {
		t.Error("应有以app为前缀的单词")
	}
	if trie.StartsWith("apx") {
		t.Error("不应有以apx为前缀的单词")
	}

	// 插入另一个单词
	trie.Insert("app")
	if !trie.Search("app") {
		t.Error("应能搜索到已插入的app")
	}
	if !trie.StartsWith("app") {
		t.Error("应有以app为前缀的单词")
	}
}

// 测试删除功能
func TestTrie_DeleteOperations(t *testing.T) {
	trie := NewTrieTree()
	words := []string{"apple", "app", "application", "banana", "ball", "cat"}

	// 插入所有测试单词
	for _, word := range words {
		trie.Insert(word)
	}

	// 测试删除不存在的单词（不应报错）
	trie.Delete("nonexistent")

	// 测试删除独立单词
	trie.Delete("cat")
	if trie.Search("cat") {
		t.Error("删除后不应再搜索到cat")
	}
	if !trie.Search("apple") {
		t.Error("删除cat不应影响apple")
	}

	// 测试删除作为其他单词前缀的单词
	trie.Delete("app")
	if trie.Search("app") {
		t.Error("删除后不应再搜索到app")
	}
	if !trie.Search("apple") {
		t.Error("删除app不应影响apple")
	}
	if !trie.StartsWith("app") {
		t.Error("删除app后仍应有以app为前缀的单词")
	}

	// 测试删除有子节点的单词
	trie.Delete("apple")
	if trie.Search("apple") {
		t.Error("删除后不应再搜索到apple")
	}
	if !trie.Search("application") {
		t.Error("删除apple不应影响application")
	}

	// 测试删除叶子节点单词
	trie.Delete("application")
	if trie.Search("application") {
		t.Error("删除后不应再搜索到application")
	}
	if trie.StartsWith("app") {
		t.Error("所有以app为前缀的单词删除后，前缀匹配应返回false")
	}

	// 测试删除另一个分支的单词
	trie.Delete("banana")
	if trie.Search("banana") {
		t.Error("删除后不应再搜索到banana")
	}
	if !trie.Search("ball") {
		t.Error("删除banana不应影响ball")
	}
}

// 测试边界情况
func TestTrie_EdgeCases(t *testing.T) {
	trie := NewTrieTree()

	// 测试空字符串
	trie.Insert("")
	if !trie.Search("") {
		t.Error("应能搜索到空字符串")
	}
	trie.Delete("")
	if trie.Search("") {
		t.Error("删除后不应再搜索到空字符串")
	}

	// 测试重复插入
	trie.Insert("test")
	trie.Insert("test") // 重复插入
	if !trie.Search("test") {
		t.Error("重复插入后应能搜索到test")
	}
	trie.Delete("test")
	if trie.Search("test") {
		t.Error("删除后不应再搜索到test")
	}

	// 测试单个字符
	trie.Insert("a")
	if !trie.Search("a") {
		t.Error("应能搜索到单个字符a")
	}
	if !trie.StartsWith("a") {
		t.Error("前缀匹配a应返回true")
	}
	trie.Delete("a")
	if trie.Search("a") {
		t.Error("删除后不应再搜索到a")
	}
	if trie.StartsWith("a") {
		t.Error("删除a后前缀匹配a应返回false")
	}

	// 测试长单词
	longWord := "abcdefghijklmnopqrstuvwxyz"
	trie.Insert(longWord)
	if !trie.Search(longWord) {
		t.Error("应能搜索到长单词")
	}
	if !trie.StartsWith("abcde") {
		t.Error("前缀匹配abcde应返回true")
	}
	trie.Delete(longWord)
	if trie.Search(longWord) {
		t.Error("删除后不应再搜索到长单词")
	}
}
