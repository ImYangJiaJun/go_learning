package errhandling

import (
	"errors"
	"fmt"
)

type UserService struct {
	store Store
}

func NewUserService(s Store) UserService {
	return UserService{
		store: s,
	}
}

// FindUser 是整个练习的核心功能，语义只有三条：
//   - 找到 → 返回 User 和 nil
//   - 底层返回 ErrUserNotFound → 原样返回（不包装）。
//     返回 err 本身而不是 ErrUserNotFound 字面量：若底层给哨兵包过上下文，原样返回才不会丢。
//   - 底层返回其他错误 → %w 包装：新文案带上 id（上下文不丢），原始错误身份留在错误链上（可用 errors.Is 认出）
func (s UserService) FindUser(id int) (User, error) {
	user, err := s.store.Get(id)
	if err == nil {
		return user, nil
	}
	if errors.Is(err, ErrUserNotFound) {
		return User{}, err
	}
	return User{}, fmt.Errorf("find user %d: %w", id, err)
}
