package main

type TransactionStatus string

const (
	TransactionStatusPending TransactionStatus = "pending"
	TransactionStatusPaid    TransactionStatus = "paid"
	TransactionStatusSuccess TransactionStatus = "success"
	TransactionStatusFailed  TransactionStatus = "failed"
)

var TransactionStatusList = []TransactionStatus{
	TransactionStatusPending,
	TransactionStatusPaid,
	TransactionStatusSuccess,
	TransactionStatusFailed,
}

type PaymentMethod string

const (
	PaymentMethodWallet PaymentMethod = "wallet"
	PaymentMethodCash   PaymentMethod = "cash"
)

var PaymentMethodList = []PaymentMethod{
	PaymentMethodWallet,
	PaymentMethodCash,
}

type Transaction struct {
	ID          int64
	Code        string
	Status      TransactionStatus
	PaymentType PaymentMethod
}
