package wallet

import (
	"fmt"
	"github.com/SonnLarissa/wallet/pkg/types"
	"github.com/google/uuid"
	"log"
	"reflect"
	"testing"
)

type testService struct {
	*Service
}

func newTestService() *testService {
	return &testService{Service: &Service{}}
}

func (s *testService) addAccountWithBalance(phone types.Phone, balance types.Money) (*types.Account, error) {
	account, err := s.RegisterAccount(phone)
	if err != nil {
		return nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, balance)
	if err != nil {
		return nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	return account, nil
}

type testAccount struct {
	phone    types.Phone
	balance  types.Money
	payments []struct {
		amount   types.Money
		category types.PaymentCategory
	}
}

func (s *testService) addAccount(data testAccount) (*types.Account, []*types.Payment, error) {
	account, err := s.RegisterAccount(data.phone)
	if err != nil {
		return nil, nil, fmt.Errorf("can't register account, error = %v", err)
	}
	err = s.Deposit(account.ID, data.balance)
	if err != nil {
		return nil, nil, fmt.Errorf("can't deposit account, error = %v", err)
	}
	payments := make([]*types.Payment, len(data.payments))
	for i, payment := range data.payments {
		payments[i], err = s.Pay(account.ID, payment.amount, payment.category)
		if err != nil {
			return nil, nil, fmt.Errorf("can't make payment, error = %v", err)
		}
	}
	return account, payments, nil
}

var defaultTestAccount = testAccount{
	phone:   "+992925556644",
	balance: 10_000_00,
	payments: []struct {
		amount   types.Money
		category types.PaymentCategory
	}{
		{amount: 1_000_00, category: "auto"},
	},
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't create payment, error = %v", err)
		return
	}
	payment := payments[0]
	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): error = %v", err)
		return
	}
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned = %v", err)
		return
	}
}

func TestService_FindPaymentByID_fail(t *testing.T) {
	s := newTestService()
	_, _, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't create payment, error = %v", err)
		return
	}
	_, err = s.FindPaymentByID(uuid.New().String())
	if err == nil {
		t.Error("FindPaymentByID(): must return error, returned nil")
		return
	}
	if err != ErrPaymentNotFound {
		t.Errorf("FindPaymentByID(): must return ErrPaymentNotFound, returned = %v", err)
		return
	}
}

func TestService_Reject_success(t *testing.T) {
	s := newTestService()
	_, payments, err := s.addAccount(defaultTestAccount)
	if err != nil {
		t.Error(err)
		return
	}
	payment := payments[0]
	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): error = %v", err)
		return
	}
	savedPayment, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't find payment by id, error =%v", err)
		return
	}
	if savedPayment.Status != types.PaymentStatusFail {
		t.Errorf("Reject(): status didn't changed, payment = %v", savedPayment)
		return
	}
	savedAccount, err := s.FindAccountByID(payment.AccountID)
	if err != nil {
		t.Errorf("Reject(): can't find account by id, error = %v", err)
		return
	}
	if savedAccount.Balance != defaultTestAccount.balance {
		t.Errorf("Reject(): balance didn't changed, account = %v", savedAccount)
		return
	}

}

func TestService_Repeat_success(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
		payments: make([]*types.Payment, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	p, _ := srv.Repeat(pp.ID)
	p.ID = pp.ID
	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Repeat(): expected %v returned = %v", pp, p)
	}
}

func TestService_Repeat_fail(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
		payments: make([]*types.Payment, 0),
	}
	_, _ = srv.RegisterAccount("+992928885522")

	_, err := srv.Repeat(uuid.New().String())
	if err == nil {
		t.Error("Repeat(): must return error, returned nil")
		return
	}
}

func TestService_FavoritePayment_success(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	_, err := srv.FavoritePayment(pp.ID, "sidal")

	if err != nil {
		t.Error("FavoritePayment(): can't make favorite return error, returned nil")
		return
	}
}

func TestService_FavoritePayment_fail(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	_, err := srv.FavoritePayment(pp.ID, "sidal")
	_, err = srv.FavoritePayment(pp.ID, "sidal")

	if err == nil {
		t.Error("FavoritePayment(): must return error, returned nil")
		return
	}
}

func TestService_PayFromFavorite_success(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	fw, _ := srv.FavoritePayment(pp.ID, "sidal")

	_, err := srv.PayFromFavorite(fw.ID)
	if err != nil {
		t.Error("PayFromFavorite(): can't make favorite return error, returned nil")
		return
	}
}

func TestService_PayFromFavorite_fail(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, err := srv.PayFromFavorite(uuid.New().String())
	if err == nil {
		t.Error("FavoritePayment(): must return error, returned nil")
		return
	}
}

func TestService_FindFavoriteByID_success(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	fw, _ := srv.FavoritePayment(pp.ID, "sidal")

	_, err := srv.FindFavoriteByID(fw.ID)
	if err != nil {
		t.Error("FindFavoriteByID(): can't make favorite return error, returned nil")
	}
}

func TestService_FindFavoriteByID_fail(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, err := srv.FindFavoriteByID(uuid.New().String())

	if err == nil {
		t.Error("FindFavoriteByID(): must return error, returned nil")
	}
}

func TestService_Export(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac, _ := srv.RegisterAccount("+992928885522")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	fw, _ := srv.FavoritePayment(pp.ID, "sidal")

	_, err := srv.PayFromFavorite(fw.ID)
	err = srv.Export("./")
	if err != nil {
		t.Error("test")
	}
}

func BenchmarkService_SumPayments(b *testing.B) {
	var s Service

	ac, err := s.RegisterAccount("+99292929292")
	if err != nil {
		b.Errorf("registerAccount return error, account => %v", ac)
	}

	err = s.Deposit(ac.ID, 100_00)
	if err != nil {
		b.Errorf("deposit return error, %v", err)
	}
	_, err = s.Pay(ac.ID, 1, "auto")
	_, err = s.Pay(ac.ID, 2, "auto")
	_, err = s.Pay(ac.ID, 3, "auto")
	_, err = s.Pay(ac.ID, 4, "auto")
	_, err = s.Pay(ac.ID, 5, "auto")
	_, err = s.Pay(ac.ID, 6, "auto")
	_, err = s.Pay(ac.ID, 7, "auto")
	_, err = s.Pay(ac.ID, 8, "auto")
	_, err = s.Pay(ac.ID, 9, "auto")
	_, err = s.Pay(ac.ID, 10, "auto")
	_, err = s.Pay(ac.ID, 11, "auto")
	if err != nil {
		b.Errorf("pay return error, %v", err)
	}
	want := types.Money(66)
	got := s.SumPayments(2)

	if want != got {
		b.Errorf("error, want =>%v, got => %v", want, got)
	}
}

func TestService_ExportToFile(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, _ = srv.RegisterAccount("+992928885522")
	_, _ = srv.RegisterAccount("+992928000000")
	_, _ = srv.RegisterAccount("+992928811111")
	err := srv.ExportToFile("salom.txt")
	println(err)
}

func TestService_ImportFromFile(t *testing.T) {
	srv := &Service{accounts: make([]*types.Account, 0)}
	err := srv.ImportFromFile("salom.txt")
	println(err)
}

func TestService_Export_2(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_, _ = srv.RegisterAccount("+992928885522")
	_, _ = srv.RegisterAccount("+992928000000")
	ac, _ := srv.RegisterAccount("+992928811111")
	_ = srv.Deposit(ac.ID, 500)

	pp, _ := srv.Pay(ac.ID, 5, "salom")

	_, _ = srv.FavoritePayment(pp.ID, "sidal")
	err := srv.Export("./data")
	exp := srv.accounts
	srv.accounts = append(srv.accounts[0:1], srv.accounts[2:]...)
	//srv.payments = make([]*types.Payment, 0)
	//srv.favorites = make([]*types.Favorite, 0)

	println(exp)
	err = srv.Import("./data")
	err = srv.Export("./data1")
	if err != nil {
		panic(err)
	}
	println(err)
}

func BenchmarkSumPayments(b *testing.B) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	account, err := srv.RegisterAccount("+992926574322")
	if err != nil {
		b.Errorf("account => %v", account)
	}
	err = srv.Deposit(account.ID, 100_00)
	if err != nil {
		b.Errorf("error => %v", err)
	}
	want := types.Money(55)
	for i := types.Money(1); i <= 10; i++ {
		_, err := srv.Pay(account.ID, i, "aa")
		if err != nil {
			b.Errorf("error => %v", err)
		}
	}
	got := srv.SumPayments(5)
	if want != got {
		b.Errorf("want => %v got => %v", want, got)
	}
}

func BenchmarkFilterPayments(b *testing.B) {
	s := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	ac1, err := s.RegisterAccount("+921111110")
	ac2, err := s.RegisterAccount("+921111111")
	ac3, err := s.RegisterAccount("+921111112")
	ac4, err := s.RegisterAccount("+921111113")
	ac5, err := s.RegisterAccount("+921111114")
	ac6, err := s.RegisterAccount("+921111115")
	if err != nil {
		return
	}
	err = s.Deposit(ac6.ID, 30_000)
	err = s.Deposit(ac5.ID, 30_000)
	err = s.Deposit(ac1.ID, 30_000)
	err = s.Deposit(ac2.ID, 30_000)
	err = s.Deposit(ac3.ID, 30_000)
	err = s.Deposit(ac4.ID, 30_000)
	err = s.Deposit(ac6.ID, 30_000)
	err = s.Deposit(ac1.ID, 30_000)
	err = s.Deposit(ac6.ID, 30_000)
	if err != nil {
		return
	}

	s.Pay(ac1.ID, 20_000, "auto")
	s.Pay(ac1.ID, 20_000, "auto")
	s.Pay(ac2.ID, 20_000, "auto")
	s.Pay(ac2.ID, 20_000, "auto")
	s.Pay(ac2.ID, 20_000, "auto")
	s.Pay(ac3.ID, 20_000, "auto")
	s.Pay(ac3.ID, 20_000, "auto")
	s.Pay(ac3.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac5.ID, 20_000, "auto")
	s.Pay(ac5.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")

	account, err := s.FilterPayments(ac6.ID, 4)
	if err != nil {
		b.Errorf("account not found ===>")
	}
	log.Println(len(account))
}
func BenchmarkService_FilterPaymentsByFn(b *testing.B) {
	s := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	filter := func(payment types.Payment) bool {
		for _, value := range s.payments {
			if payment.ID == value.ID {
				return true
			}
		}
		return false
	}
	ac1, err := s.RegisterAccount("+921111110")
	ac2, err := s.RegisterAccount("+921111111")
	ac3, err := s.RegisterAccount("+921111112")
	ac4, err := s.RegisterAccount("+921111113")
	ac5, err := s.RegisterAccount("+921111114")
	ac6, err := s.RegisterAccount("+921111115")
	if err != nil {
		return
	}
	err = s.Deposit(ac6.ID, 30_000)
	err = s.Deposit(ac5.ID, 30_000)
	err = s.Deposit(ac1.ID, 30_000)
	err = s.Deposit(ac2.ID, 30_000)
	err = s.Deposit(ac3.ID, 30_000)
	err = s.Deposit(ac4.ID, 30_000)
	err = s.Deposit(ac6.ID, 30_000)
	err = s.Deposit(ac1.ID, 30_000)
	err = s.Deposit(ac6.ID, 30_000)
	if err != nil {
		return
	}

	s.Pay(ac1.ID, 20_000, "auto")
	s.Pay(ac1.ID, 20_000, "auto")
	s.Pay(ac2.ID, 20_000, "auto")
	s.Pay(ac2.ID, 20_000, "auto")
	s.Pay(ac2.ID, 20_000, "auto")
	s.Pay(ac3.ID, 20_000, "auto")
	s.Pay(ac3.ID, 20_000, "auto")
	s.Pay(ac3.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac4.ID, 20_000, "auto")
	s.Pay(ac5.ID, 20_000, "auto")
	s.Pay(ac5.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")
	s.Pay(ac6.ID, 20_000, "auto")

	account, err := s.FilterPaymentsByFn(filter, 4)
	if err != nil {
		b.Errorf("account not found ===>")
	}
	log.Println(len(account))
}
