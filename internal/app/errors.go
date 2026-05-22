package app

import "errors"

var (
	ErrInitTokenMaker    = errors.New("init token maker")
	ErrInitUserModule    = errors.New("init user module")
	ErrInitAuthModule    = errors.New("init auth module")
	ErrInitCartModule    = errors.New("init cart module")
	ErrInitProductModule = errors.New("init product module")
	ErrInitPaymentModule = errors.New("init payment module")
)
