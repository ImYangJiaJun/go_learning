package errhandling

import "errors"

type User struct {
	ID   int
	Name string
	Age  int
}

// 哨兵错误：包级变量，调用方用 errors.Is 匹配
var ErrUserNotFound = errors.New("user not found")
var ErrStorage = errors.New("storage failure")

// Store 存储层抽象：让"底层故障"在测试中可用替身确定性地触发
type Store interface {
	Get(id int) (User, error)
}

// MapStore 是 Store 的 map 实现，唯一字段就是那张 map
type MapStore struct {
	users map[int]User
}

func NewMapStore(users map[int]User) MapStore {
	return MapStore{users: users}
}

// Get：id 存在 → 返回 User 和 nil；不存在 → 返回 User{} 和 ErrUserNotFound
func (m MapStore) Get(id int) (User, error) {
	user, ok := m.users[id]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}
