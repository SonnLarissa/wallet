package wallet

import (
	"errors"
	"github.com/google/uuid"
	"github.com/SonnLarissa/wallet/pkg/types"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	ErrPhoneRegistered      = errors.New("phone already registered")
	ErrFavoriteRegistered   = errors.New("favorite already registered")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrAccountNotFound      = errors.New("account not found")
	ErrNotEnoughBalance     = errors.New("not enough balance")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrFavoriteNotFound     = errors.New("favorite not found")
)

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

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
		}
	}

	if account == nil {
		return ErrAccountNotFound
	}
	account.Balance += amount
	return nil
}

func (s *Service) Pay(accountID int64, amount types.Money, category types.PaymentCategory) (*types.Payment, error) {
	if amount <= 0 {
		return nil, ErrAmountMustBePositive
	}
	var account *types.Account
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			account = acc
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
		AccountID: accountID,
		Amount:    amount,
		Category:  category,
		Status:    types.PaymentStatusInProgress,
	}
	s.payments = append(s.payments, payment)
	return payment, nil
}

func (s *Service) Reject(paymentID string) error {
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return err
	}
	account, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		return err
	}
	payment.Status = types.PaymentStatusFail
	account.Balance += payment.Amount
	return nil
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	p, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	pp, err := s.Pay(p.AccountID, p.Amount, p.Category)
	if err != nil {
		return nil, err
	}
	return pp, nil
}

func (s *Service) FavoritePayment(paymentID string, name string) (*types.Favorite, error) {
	for _, favorite := range s.favorites {
		if favorite.Name == name {
			return nil, ErrFavoriteRegistered
		}
	}
	payment, err := s.FindPaymentByID(paymentID)
	if err != nil {
		return nil, err
	}
	s.nextAccountID++
	favorite := &types.Favorite{
		ID:        uuid.New().String(),
		AccountID: payment.AccountID,
		Name:      name,
		Amount:    payment.Amount,
		Category:  payment.Category,
	}
	s.favorites = append(s.favorites, favorite)
	return favorite, nil
}

func (s *Service) PayFromFavorite(favoriteID string) (*types.Payment, error) {
	fw, err := s.FindFavoriteByID(favoriteID)
	if err != nil {
		return nil, err
	}
	return s.Pay(fw.AccountID, fw.Amount, fw.Category)
}

func (s *Service) FindAccountByID(accountID int64) (*types.Account, error) {
	for _, acc := range s.accounts {
		if acc.ID == accountID {
			return acc, nil
		}
	}
	return nil, ErrAccountNotFound
}

func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, py := range s.payments {
		if py.ID == paymentID {
			return py, nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, py := range s.favorites {
		if py.ID == favoriteID {
			return py, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

func (s *Service) ExportToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, account := range s.accounts {
		_, err = file.Write([]byte(account.ToString() + "|"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ImportFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	str, _ := io.ReadAll(file)
	arr := strings.Split(string(str), "|")
	for _, ac := range arr {
		accountStr := strings.Split(ac, ";")
		if len(accountStr) < 2 {
			continue
		}
		account, err := s.RegisterAccount(types.Phone(accountStr[1]))
		if err != nil {
			return err
		}
		m, _ := strconv.Atoi(accountStr[2])
		err = s.Deposit(account.ID, types.Money(m))
		if err != nil {
			return err
		}
	}
	return nil
}