package wallet

import (
	"errors"
	"fmt"
	"testing"
)

// 行为 1：新钱包余额为 0，存款后余额增加
func TestDeposit(t *testing.T) {
	wallet := Wallet{}
	assertBalance(t, wallet, Bitcoin(0))

	wallet.Deposit(Bitcoin(10))
	assertBalance(t, wallet, Bitcoin(10))
}

// 行为 2：取款成功扣减余额；Stringer 让失败输出说人话
func TestWithdraw(t *testing.T) {
	t.Run("先存后取", func(t *testing.T) {
		wallet := Wallet{}
		wallet.Deposit(Bitcoin(10))

		err := wallet.Withdraw(Bitcoin(5))
		if err != nil {
			t.Fatalf("不该返回错误，却得到: %v", err)
		}
		assertBalance(t, wallet, Bitcoin(5))
	})

	t.Run("Stringer 打印说人话", func(t *testing.T) {
		got := fmt.Sprintf("%s", Bitcoin(10))
		want := "10 BTC"
		if got != want {
			t.Errorf("得到 %q，期望 %q", got, want)
		}
	})
}

// 行为 3：余额不足时取款失败且余额一分不动；刚好取光允许
func TestWithdrawInsufficientFunds(t *testing.T) {
	t.Run("余额不足", func(t *testing.T) {
		wallet := Wallet{}
		wallet.Deposit(Bitcoin(10))

		err := wallet.Withdraw(Bitcoin(100))
		assertError(t, err, ErrInsufficientFunds)
		assertBalance(t, wallet, Bitcoin(10))
	})

	t.Run("刚好取光", func(t *testing.T) {
		wallet := Wallet{}
		wallet.Deposit(Bitcoin(10))

		err := wallet.Withdraw(Bitcoin(10))
		if err != nil {
			t.Fatalf("不该返回错误，却得到: %v", err)
		}
		assertBalance(t, wallet, Bitcoin(0))
	})
}

func assertBalance(t testing.TB, wallet Wallet, want Bitcoin) {
	t.Helper()
	got := wallet.Balance()
	if got != want {
		// %s 触发 Stringer 回调，失败输出显示 "10 BTC" 而不是 "10"
		t.Errorf("得到 %s，期望 %s", got, want)
	}
}

func assertError(t testing.TB, err error, want error) {
	t.Helper()
	if err == nil {
		t.Fatal("期望返回错误，却没有返回")
	}
	if !errors.Is(err, want) {
		t.Errorf("得到 %v，期望 %v", err, want)
	}
}
