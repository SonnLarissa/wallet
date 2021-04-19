package wallet

import (
	"errors"
	"github.com/SonnLarissa/wallet/pkg/types"
	"github.com/google/uuid"
	"reflect"
	"testing"
)

func TestService_FindAccountByID_success(t *testing.T) {
	svc := &Service{accounts: []*types.Account{
		{ID: 1, Phone: "+99295864521", Balance: 10},
		{ID: 2, Phone: "+99295121456", Balance: 10},
		{ID: 3, Phone: "+99295864512", Balance: 10},
	}}
	account, _ := svc.FindAccountByID(2)

	accountExp := svc.accounts[1]

	if !reflect.DeepEqual(accountExp, account) {
		t.Errorf("Invalid result, expected : %v, actual %v", accountExp, account)
	}
}
func TestService_FindAccountByID_notFound(t *testing.T) {
	svc := &Service{accounts: []*types.Account{
		{ID: 1, Phone: "+99295864521", Balance: 10},
		{ID: 2, Phone: "+99295121456", Balance: 10},
		{ID: 3, Phone: "+99295864512", Balance: 10},
	}}
	_, error := svc.FindAccountByID(4)

	err := errors.New("account not found")

	if !reflect.DeepEqual(error, err) {
		t.Errorf("Invalid result, expected : %v, actual %v", err, error)
	}
}

func TestService_Reject_success(t *testing.T) {
	s := &Service{}
	phone := types.Phone("+992000000001")
	account, err := s.RegisterAccount(phone)
	if err != nil {
		t.Errorf("Reject(): can't register account, error = %v", err)
		return
	}
	s.Deposit(account.ID, 10000_00)
	if err != nil {
		t.Errorf("Reject(): can't deposit  account, error = %v", err)
		return
	}

	payment, err := s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("Reject(): can't create payment, error = %v", err)
		return
	}

	err = s.Reject(payment.ID)
	if err != nil {
		t.Errorf("Reject(): can't reject payment, error = %v", err)
		return
	}
}
func TestService_Reject_failed(t *testing.T) {
	s := &Service{}
	phone := types.Phone("+992000000001")
	account, err := s.RegisterAccount(phone)
	if err != nil {
		t.Errorf("Reject(): can't register account, error = %v", err)
		return
	}
	s.Deposit(account.ID, 10000_00)

	_, err = s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("Reject(): can't create payment, error = %v", err)
		return
	}

	var fakeId string = "2"
	err = s.Reject(fakeId)
	if err == nil {
		t.Errorf("Reject(): can't reject payment, error = %v", err)
		return
	}
}

func TestService_FindPaymentByID_success(t *testing.T) {
	s := &Service{}
	phone := types.Phone("+992000000001")
	account, err := s.RegisterAccount(phone)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't register account, error = %v", err)
		return
	}
	s.Deposit(account.ID, 10000_00)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't deposit  account, error = %v", err)
		return
	}

	payment, err := s.Pay(account.ID, 1000_00, "auto")
	if err != nil {
		t.Errorf("FindPaymentByID(): can't create payment, error = %v", err)
		return
	}

	got, err := s.FindPaymentByID(payment.ID)
	if err != nil {
		t.Errorf("FindPaymentByID(): can't reject payment, error = %v", err)
		return
	}
	if !reflect.DeepEqual(payment, got) {
		t.Errorf("FindPaymentByID(): wrong payment returned, error = %v", err)
		return
	}
}

func TestService_Repeat_success(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
		payments: make([]*types.Payment, 0),
	}
	ac, _ := srv.RegisterAccount("+992927808989")
	_ = srv.Deposit(ac.ID, 500)
	pp, _ := srv.Pay(ac.ID, 5, "auto")
	p, _ := srv.Repeat(pp.ID)
	p.ID = pp.ID
	if !reflect.DeepEqual(p, pp) {
		t.Errorf("Repeat(): expected %v returned =%v", pp, p)
	}
}

func TestService_Repeat_failed(t *testing.T) {
	srv := &Service{
		accounts: make([]*types.Account, 0),
		payments: make([]*types.Payment, 0),
	}
	_, _ = srv.RegisterAccount("+99298786545")
	_, err := srv.Repeat(uuid.New().String())
	if err == nil {
		t.Errorf("Repeat(): must retrun error, returned nil")
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
func TestService_ExportToFile(t *testing.T) {
	srv := &Service{
		accounts:  make([]*types.Account, 0),
		payments:  make([]*types.Payment, 0),
		favorites: make([]*types.Favorite, 0),
	}
	_,_ = srv.RegisterAccount("+992928885522")
	_,_ = srv.RegisterAccount("+992928000000")
	_,_ = srv.RegisterAccount("+992928811111")
	err := srv.ExportToFile("salom.txt")
	println(err)
}


func TestService_ImportFromFile(t *testing.T) {
	srv := &Service{accounts: make([]*types.Account, 0)}
	err := srv.ImportFromFile("salom.txt")
	println(err)
}
