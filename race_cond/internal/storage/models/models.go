package models

type User struct {
	Id       int64
	Login    string
	Password string
	Balance  int
	GotBonus int
}
