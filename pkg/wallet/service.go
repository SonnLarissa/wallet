package wallet

import (
	"errors"
	"github.com/SonnLarissa/wallet/pkg/types"
	"github.com/google/uuid"
)

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
}

var (
	ErrPhoneRegistered      = errors.New("phone already registered")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrAccountNotFound      = errors.New("account not found")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrNotEnoughBalance     = errors.New("mot enough balance")
)

func (s *Service) RegisterAccount(phone types.Phone) (*types.Account, error) {
	for _, account := range s.accounts {
		if account.Phone == phone {
			return nil, ErrPhoneRegistered
		}
	}
	s.nextAccountID++
	account := &types.Account{
		ID:      s.nextAccountID,
		Phone:   phone,
		Balance: 0,
	}
	s.accounts = append(s.accounts, account)
	return account, nil
}

func (s *Service) Deposit(accountID int64, amount types.Money) error {
	if amount <= 0 {
		return ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
		if account == nil {
			return ErrAccountNotFound
		}
	}
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	if account.Balance < amount {
		return nil, ErrNotEnoughBalance
	}
	account.Balance -= amount
	paymentID := uuid.New().String()
	payment := &types.Payment{
		ID:        paymentID,
		AccountID: accID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
			break
		}
	}
	if account == nil {
		return nil, ErrAccountNotFound
	}
	return account, nil
}



func (s *Service) Reject(paymentID string) error {
	var target *types.Payment
	for _, payment := range s.payments {
		if payment.ID == paymentID {
			target = payment
			break
		}
	}
	if target == nil {
		return ErrPaymentNotFound
	}
	var targetAccount *types.Account
	for _, acc := range s.accounts {
		if acc.ID == target.AccountID {
			targetAccount = acc
			break
		}
	}
	if targetAccount == nil {
		return ErrAccountNotFound
	}
	target.Status = types.PaymentStatusFail
	targetAccount.Balance += target.Amount
	return nil
}


//  FindPaymentByID возвраûает указателþ на найденнýй платёж и nil в каùестве
//оúибки
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, paym := range s.payments {
		if paym.ID == paymentID {
			return paym, nil
		}
	}
	return nil, ErrPaymentNotFound
}
