package app

import "errors"

var (
	ErrInitTokenMaker     = errors.New("init token maker")
	ErrInitUserModule     = errors.New("init user module")
	ErrInitAuthModule     = errors.New("init auth module")
	ErrInitCartModule     = errors.New("init cart module")
	ErrInitOrderModule    = errors.New("init order module")
	ErrInitCheckoutModule = errors.New("init checkout module")
	ErrInitProductModule  = errors.New("init product module")
	ErrInitPaymentModule  = errors.New("init payment module")
)
