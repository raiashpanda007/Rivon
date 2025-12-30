package types

type RegisterType struct {
	Name     string `json:"name" validator:"required"`
	Password string `json:"password" validator:"required"`
	Email    string `json:"email" validator:"required"`
}

type LoginType struct {
	Email    string `json:"email" validator:"required"`
	Password string `json:"password" validator:"required"`
}

type RefreshTokenType struct {
	Id string `json:"id" validator:"required"`
}
type VerifyOTPCredentials struct {
	OTP string `json:"otp" validator:"required"`
}

type TransactionType string

const (
	DEBIT  TransactionType = "debit"
	CREDIT TransactionType = "credit"
)
