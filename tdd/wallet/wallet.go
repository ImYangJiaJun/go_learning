package wallet

import (
	"errors"
	"fmt"
)

// Bitcoin 自定义类型，底层是 int，单位"枚"。
// 注意不是别名：Bitcoin(10) 与 int(10) 在编译期就是两种东西，
// 编译器会挡住"把普通数字当钱存"这类错误。
type Bitcoin int

// String 让 Bitcoin 实现 fmt.Stringer 接口：
// 之后 %v/%s 打印（包括 t.Errorf 的失败输出）显示 "10 BTC" 而不是 "10"。
// 必须是值接收者；实现内部用 %d 格式化（原因见行为 2 知识点）。
func (b Bitcoin) String() string {
	str := fmt.Sprintf("%d BTC", b)
	return str
}

// Wallet 比特币钱包。唯一字段 balance 不导出——余额只能经 Deposit/Withdraw
// 改变，调用方无法直接改字段绕开规则。零值即可用，没有构造器：测试直接写 Wallet{}。
type Wallet struct {
	balance Bitcoin
}

// Deposit 存入 amount。必须指针接收者：改的是原钱包，不是副本。
func (w *Wallet) Deposit(amount Bitcoin) {
	w.balance += amount
}

// Balance 返回当前余额。契约统一用指针接收者（原因见行为 1 知识点）。
func (w *Wallet) Balance() Bitcoin {
	return w.balance
}

// Withdraw 取出 amount：
//   - 余额充足（含刚好取光）→ 扣款，返回 nil
//   - 余额不足 → 余额一分不动，返回 ErrInsufficientFunds
func (w *Wallet) Withdraw(amount Bitcoin) error {
	if w.balance < amount {
		return ErrInsufficientFunds
	}
	w.balance -= amount
	return nil
}

// 哨兵错误：包级变量，全进程唯一实例；调用方用 errors.Is 识别。
var ErrInsufficientFunds = errors.New("cannot withdraw, insufficient funds")
