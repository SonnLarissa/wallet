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
	"sync"
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
		for _, favorite := range s.favorites {
			data.WriteString(favorite.ToString() + "\n")
		}
		err := save(data.String(), "favorites")
		if err != nil {
			return err
		}
	}

	if len(s.payments) > 0 {
		data := strings.Builder{}
		for _, payment := range s.payments {
			data.WriteString(payment.ToString() + "\n")
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

func (s *Service) ExportAccountHistory(accountID int64) ([]types.Payment, error) {

	account, err := s.FindAccountByID(accountID)

	if err != nil {
		return nil, err
	}

	var payments []types.Payment
	for _, v := range s.payments {
		if v.AccountID == account.ID {
			data := types.Payment{
				ID:        v.ID,
				AccountID: v.AccountID,
				Amount:    v.Amount,
				Category:  v.Category,
				Status:    v.Status,
			}
			payments = append(payments, data)
		}
	}
	return payments, nil
}

func (s *Service) HistoryToFiles(payments []types.Payment, dir string, records int) error {

	if len(payments) > 0 {
		if len(payments) <= records {
			file, _ := os.OpenFile(dir+"/payments.dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
			defer file.Close()

			var str string
			for _, v := range payments {
				str += v.ToString() + "\n"
			}
			_, _ = file.WriteString(str)
		} else {

			var str string
			k := 0
			t := 1
			var file *os.File
			for _, v := range payments {
				if k == 0 {
					file, _ = os.OpenFile(dir+"/payments"+fmt.Sprint(t)+".dump", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
				}
				k++
				str = v.ToString() + "\n"
				_, _ = file.WriteString(str)
				if k == records {
					str = ""
					t++
					k = 0
					file.Close()
				}
			}

		}
	}

	return nil
}

func (s *Service) SumPayments(goroutines int) types.Money {

	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	len1 := len(s.payments) / goroutines
	len2 := len(s.payments) % goroutines
	sum := types.Money(0)
	if goroutines > 1 {
		currentPosition := int(0)
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			lastPosition := currentPosition

			go func(lastPosition int) {
				defer wg.Done()
				res := types.Money(0)
				curPaymArray := s.payments[lastPosition : lastPosition+len1]
				for _, currentPayment := range curPaymArray {
					if currentPayment == nil {
						continue
					}
					res += currentPayment.Amount
				}
				mu.Lock()
				mu.Unlock()
				sum += res
			}(lastPosition)

			currentPosition += len1

		}
	} else {
		for _, currentPayment := range s.payments {
			sum += currentPayment.Amount
		}
	}
	if len2 != 0 {
		for _, curPayment := range s.payments[(len(s.payments) - len2):len(s.payments)] {
			sum += curPayment.Amount
		}
	}

	wg.Wait()
	return sum

}

func (s *Service) FilterPayments(accountID int64, goroutines int) (resPayments []types.Payment, err error) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	len1 := len(s.payments) / goroutines
	len2 := len(s.payments) % goroutines

	if goroutines < 2 {
		for _, res := range s.payments {
			if res.AccountID == accountID {
				resPayments = append(resPayments, *res)
			}
			if res == nil {
				return nil, ErrAccountNotFound
			}
		}
		return
	}

	currentPosition := int(0)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		lastPosition := currentPosition
		go func(lastPosition int) {
			defer wg.Done()
			res := []types.Payment{}
			curPaymArray := s.payments[lastPosition : lastPosition+len1]

			for _, currentPayment := range curPaymArray {
				if currentPayment == nil {
					continue
				}
				if currentPayment.AccountID == accountID {
					res = append(res, *currentPayment)
				}
			}
			mu.Lock()
			resPayments = append(resPayments, res...)
			mu.Unlock()
		}(lastPosition)

		currentPosition += len1
	}

	if len2 != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := []types.Payment{}
			for _, currentPayment := range s.payments[(len(s.payments) - len2):len(s.payments)] {
				if currentPayment == nil {
					continue
				}
				if currentPayment.AccountID == accountID {
					res = append(res, *currentPayment)
				}
			}

			mu.Lock()
			resPayments = append(resPayments, res...)
			mu.Unlock()

		}()
	}

	wg.Wait()
	if resPayments == nil {
		return nil, ErrAccountNotFound
	}
	return
}

func (s *Service) FilterPaymentsByFn(filter func(payment types.Payment)bool, goroutines int) (resPayments []types.Payment, err error) {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	len1 := len(s.payments) / goroutines
	len2 := len(s.payments) % goroutines

	if goroutines < 2 {
		for _, res := range s.payments {
			if filter(*res) {
				resPayments = append(resPayments, *res)
			}
			if res == nil {
				return nil, ErrAccountNotFound
			}
		}
		return
	}

	currentPosition := int(0)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		lastPosition := currentPosition
		go func(lastPosition int) {
			defer wg.Done()
			res := []types.Payment{}
			curPaymArray := s.payments[lastPosition : lastPosition+len1]

			for _, currentPayment := range curPaymArray {
				if currentPayment == nil {
					continue
				}
				if filter(*currentPayment) {
					res = append(res, *currentPayment)
				}
			}
			mu.Lock()
			resPayments = append(resPayments, res...)
			mu.Unlock()
		}(lastPosition)

		currentPosition += len1
	}

	if len2 != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := []types.Payment{}
			for _, currentPayment := range s.payments[(len(s.payments) - len2):len(s.payments)] {
				if currentPayment == nil {
					continue
				}
				if filter(*currentPayment) {
					res = append(res, *currentPayment)
				}
			}

			mu.Lock()
			resPayments = append(resPayments, res...)
			mu.Unlock()

		}()
	}

	wg.Wait()
	if resPayments == nil {
		return nil, ErrAccountNotFound
	}
	return
}
