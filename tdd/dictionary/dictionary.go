package dictionary

import "errors"

// Dictionary 单词字典：键是单词，值是释义
type Dictionary map[string]string

// 哨兵错误（包级变量）：调用方一律用 errors.Is 识别，禁止比较错误文案
var ErrNotFound = errors.New("could not find the word you were looking for")
var ErrWordExists = errors.New("cannot add word because it already exists")
var ErrWordDoesNotExist = errors.New("cannot update word because it does not exist")

// 四个方法全部是值接收者——这正是本练习的核心领悟点：
// map 是引用类型，值接收者拷贝的只是引用头，改的是同一张底层表

// Search 查词：存在 → (释义, nil)；不存在 → ("", ErrNotFound)
func (d Dictionary) Search(word string) (string, error) {
	if val, ok := d[word]; ok {
		return val, nil
	}

	return "", ErrNotFound
}

// Add 加词：词已存在 → ErrWordExists，且不覆盖旧释义；否则写入，返回 nil
func (d Dictionary) Add(word, definition string) error {
	if _, ok := d[word]; ok {
		return ErrWordExists
	}
	d[word] = definition
	return nil
}

// Update 改词：词存在 → 覆盖释义，返回 nil；不存在 → ErrWordDoesNotExist
func (d Dictionary) Update(word, definition string) error {
	if _, ok := d[word]; ok {
		d[word] = definition
		return nil
	}
	return ErrWordDoesNotExist
}

// Delete 删词：直接删除；删不存在的词是 no-op，不算错误（所以没有返回值）
func (d Dictionary) Delete(word string) {
	delete(d, word)
}
