package main

import (
	"fmt"
	"github.com/SonnLarissa/wallet/pkg/wallet"
)

func main() {
	svc := &wallet.Service{}
	account, err := svc.RegisterAccount("+992927895456")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(account)

	account, err = svc.RegisterAccount("+992927895456")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(account)

	err = svc.Deposit(account.ID, 10)
	if err != nil {
		switch err {
		case wallet.ErrAccountNotFound:
			fmt.Println("not found")
		case wallet.ErrPhoneRegistered:
			fmt.Println("already registered")
		case wallet.ErrAmountMustBePositive:
			fmt.Println("must be positive")
		}
		fmt.Println(err)
		return

	}
	fmt.Println(account.Balance) //10
	svc.RegisterAccount("+992927895456")
	svc.Deposit(1, 10)
	svc.RegisterAccount("+992231236554")

}
