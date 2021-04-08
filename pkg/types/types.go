package types

//Money представляет собой денежную сумму в мин единицах
type Money int64

//PaymentCategory представляет собой категорию. в которой был совершен платеж
type PaymentCategory string

//PaymentStatus представляет собой статус платежа
type PaymentStatus string

//Предопределнеые статусы платежей
const (
	PaymentStatusOk         PaymentStatus = "OK"
	PaymentStatusFail       PaymentStatus = "FAIL"
	PaymentStatusInProgress PaymentStatus = "INPROGRESS"
)

//Payment  представляет информацию о платеже
type Payment struct {
	ID       string
	Amount   Money
	Category PaymentCategory
	Status   PaymentStatus
}

type Phone string

//Account предаствялет информацию о счете пользоватлея
type Account struct {
	ID      int64
	Phone   Phone
	Balance Money
}
