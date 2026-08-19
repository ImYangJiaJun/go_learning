package errhandling

import (
	"errors"
	"strings"
	"testing"
)

// stubStore 测试替身：不管查什么都返回固定错误。
// MapStore 只能触发 ErrUserNotFound，模拟不了"存储层坏了"，所以行为 2 必须靠替身。
type stubStore struct {
	err error
}

func (s stubStore) Get(id int) (User, error) {
	return User{}, s.err
}

// TestFindUser 表格化测试：一行一个用例，循环里统一调用、统一断言。
// 表格的"列" = 用例之间的差异：喂什么仓库、查哪个 id、期望什么结果。
func TestFindUser(t *testing.T) {
	cases := []struct {
		name       string
		store      Store  // 每行自带仓库：真的 MapStore 或替身 stubStore，接口让它们可互换
		id         int    // 要查的用户 id
		wantUser   User   // 期望找到的用户（仅 wantErr 为 nil 时检查）
		wantErr    error  // 期望 errors.Is 命中的哨兵；nil 表示期望成功
		wantNotErr error  // 期望不误伤的哨兵（可留空）
		wantInMsg  string // 期望错误文案包含的内容（可留空，验证上下文不丢）
	}{
		{
			name:     "找到",
			store:    NewMapStore(map[int]User{1: {ID: 1, Name: "Tom", Age: 20}}),
			id:       1,
			wantUser: User{ID: 1, Name: "Tom", Age: 20},
		},
		{
			name:    "没这个人",
			store:   NewMapStore(map[int]User{}), // 空仓库，查谁都不存在
			id:      99,
			wantErr: ErrUserNotFound,
		},
		{
			name:       "存储故障",
			store:      stubStore{err: ErrStorage}, // 替身确定性触发"存储层坏了"
			id:         99,
			wantErr:    ErrStorage,      // %w 包装后 errors.Is 沿链仍能认出
			wantNotErr: ErrUserNotFound, // 故障与"没这个人"必须能区分
			wantInMsg:  "99",            // 包装文案 "find user 99: ..." 带上了 id
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := NewUserService(c.store).FindUser(c.id)

			// 期望成功的分支
			if c.wantErr == nil {
				if err != nil {
					t.Fatalf("期望没有错误，得到 %v", err)
				}
				if got != c.wantUser { // User 三个字段都可比，结构体可以直接 != 比较
					t.Errorf("期望 %+v，得到 %+v", c.wantUser, got)
				}
				return
			}

			// 期望失败的分支
			if !errors.Is(err, c.wantErr) {
				t.Errorf("期望 errors.Is 认出 %v，得到 %v", c.wantErr, err)
			}
			if c.wantNotErr != nil && errors.Is(err, c.wantNotErr) {
				t.Errorf("不应被误判为 %v，得到 %v", c.wantNotErr, err)
			}
			if c.wantInMsg != "" && !strings.Contains(err.Error(), c.wantInMsg) {
				t.Errorf("期望错误信息包含 %q，得到 %v", c.wantInMsg, err)
			}
		})
	}
}
