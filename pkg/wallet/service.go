package wallet

import (
	"errors"
	"fmt"
	"github.com/SonnLarissa/wallet/pkg/types"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Service struct {
	nextAccountID int64
	accounts      []*types.Account
	payments      []*types.Payment
	favorites     []*types.Favorite
}

var (
	ErrPhoneRegistered      = errors.New("phone already registered")
	ErrAmountMustBePositive = errors.New("amount must be greater than zero")
	ErrFavoriteRegistered   = errors.New("favorite already registered")
	ErrAccountNotFound      = errors.New("account not found")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrFavoriteNotFound     = errors.New("favorite not found")
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

//  FindPaymentByID возвращает указатель на найденый платеж
func (s *Service) FindPaymentByID(paymentID string) (*types.Payment, error) {
	for _, paym := range s.payments {
		if paym.ID == paymentID {
			return paym, nil
		}
	}
	return nil, ErrPaymentNotFound
}

func (s *Service) FindFavoriteByID(favoriteID string) (*types.Favorite, error) {
	for _, favor := range s.favorites {
		if favor.ID == favoriteID {
			return favor, nil
		}
	}
	return nil, ErrFavoriteNotFound
}

func (s *Service) Repeat(paymentID string) (*types.Payment, error) {
	res, er := s.FindPaymentByID(paymentID)
	if er != nil {
		return nil, er
	}
	res, er = s.Pay(res.AccountID, res.Amount, res.Category)
	if er != nil {
		return nil, er
	}
	return res, nil
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
	fav, er := s.FindFavoriteByID(favoriteID)
	if er != nil {
		return nil, er
	}
	return s.Pay(fav.AccountID, fav.Amount, fav.Category)
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

func (s *Service) Export(dir string) error {
	save := func(data string, name string) error {
		_ = os.Mkdir(dir, 0777)
		f, err := os.Create(dir + "/" + name + ".dump")
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(data)
		if err != nil {
			return err
		}
		return nil
	}

	if len(s.accounts) > 0 {
		data := strings.Builder{}
		for _, account := range s.accounts {
			data.WriteString(account.ToString() + "\n")
		}
		err := save(data.String(), "accounts")
		if err != nil {
			return err
		}
	}

	if len(s.favorites) > 0 {
		data := strings.Builder{}
		fmt.Println(data)
		for _, favor := range s.favorites {
			data.WriteString(favor.ToStrFavorite() + "\n")
		}
		err := save(data.String(), "favorites")
		if err != nil {
			return err
		}
	}

	if len(s.payments) > 0 {
		data := strings.Builder{}
		for _, payment := range s.payments {
			data.WriteString(payment.ToStrPayment() + "\n")
		}
		err := save(data.String(), "payments")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Import(dir string) error {
	read := func(name string) string {
		data, err := ioutil.ReadFile(dir + "/" + name + ".dump")
		if err != nil {
			return ""
		}
		return string(data)
	}

	data := read("accounts")
	arr := strings.Split(data, "\n")
	for _, ac := range arr {
		accountStr := strings.Split(ac, ";")
		if len(accountStr) < 2 {
			continue
		}
		ID, _ := strconv.Atoi(accountStr[0])
		Phone := types.Phone(accountStr[1])
		Balance, _ := strconv.Atoi(accountStr[2])
		fw, err := s.FindAccountByID(int64(ID))
		if err != nil {
			fw = &types.Account{
				ID:      int64(ID),
				Phone:   Phone,
				Balance: types.Money(Balance),
			}
			s.accounts = append(s.accounts, fw)
			s.nextAccountID = int64(ID)
		}
		fw.Phone = Phone
		fw.Balance = types.Money(Balance)
	}

	data = read("payments")
	payments := strings.Split(data, "\n")
	for _, ac := range payments {
		paymentStr := strings.Split(ac, ";")
		if len(paymentStr) < 2 {
			continue
		}
		ID := paymentStr[0]
		AccountID, _ := strconv.Atoi(paymentStr[1])
		Amount, _ := strconv.Atoi(paymentStr[2])
		Category := paymentStr[3]
		Status := paymentStr[4]
		py, err := s.FindPaymentByID(ID)
		if err == nil {
			py.AccountID = int64(AccountID)
			py.Amount = types.Money(Amount)
			py.Category = types.PaymentCategory(Category)
			py.Status = types.PaymentStatus(Status)
			continue
		}
		s.payments = append(s.payments, &types.Payment{
			ID:        ID,
			AccountID: int64(AccountID),
			Amount:    types.Money(Amount),
			Category:  types.PaymentCategory(Category),
			Status:    types.PaymentStatus(Status),
		})
	}

	data = read("favorites")
	favorites := strings.Split(data, "\n")
	for _, ac := range favorites {
		favoriteStr := strings.Split(ac, ";")
		if len(favoriteStr) < 2 {
			continue
		}
		ID := favoriteStr[0]
		AccountID, _ := strconv.Atoi(favoriteStr[1])
		Name := favoriteStr[2]
		Amount, _ := strconv.Atoi(favoriteStr[3])
		Category := favoriteStr[4]
		fw, err := s.FindFavoriteByID(ID)
		if err == nil {
			fw.AccountID = int64(AccountID)
			fw.Amount = types.Money(Amount)
			fw.Name = Name
			fw.Category = types.PaymentCategory(Category)
			continue
		}
		favorite := &types.Favorite{
			ID:        uuid.New().String(),
			AccountID: int64(AccountID),
			Amount:    types.Money(Amount),
			Name:      Name,
			Category:  types.PaymentCategory(Category),
		}
		s.favorites = append(s.favorites, favorite)
	}
	return nil
}