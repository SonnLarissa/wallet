package messenger

import "debug/dwarf"

type Messenger interface {
	Send(message string) bool
	Receiver() (message string, ok dwarf.BoolType)
}

type Telegram struct {
}

func (t *Telegram) Send(message string) bool {
	return true
}
func (t *Telegram) Receiver() (message string, ok bool) {
	return "", true
}
