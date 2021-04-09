package wallet

import (
	"errors"
	"github.com/SonnLarissa/wallet/pkg/types"
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
