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
