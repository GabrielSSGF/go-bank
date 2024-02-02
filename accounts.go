package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginResponse struct {
	Number int64  `json:"number"`
	Token  string `json:"token"`
}

type LoginRequest struct {
	Number   int64  `json:"number"`
	Password string `json:"password"`
}

type TransferRequest struct {
	ToAccount int64 `json:"to_account"`
	Amount    int   `json:"amount"`
}

type CreateAccountRequest struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	PasswordEncrypted string `json:"password"`
}

type Account struct {
	ID                int       `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Number            int64     `json:"number"`
	PasswordEncrypted string    `json:"-"`
	Balance           int64     `json:"balance"`
	CreatedAt         time.Time `json:"created_at"`
}

func (account *Account) ValidaPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(account.PasswordEncrypted), []byte(password)) == nil
}

func NewAccount(firstName, lastName, password string) (*Account, error) {
	senhaEncriptada, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return &Account{
		// ID:        rand.Intn(10000),
		FirstName:         firstName,
		LastName:          lastName,
		PasswordEncrypted: string(senhaEncriptada),
		Number:            int64(rand.Intn(100000)),
		CreatedAt:         time.Now().UTC(),
	}, nil
}
