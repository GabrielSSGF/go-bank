package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
	GetAccountByNumber(int) (*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func newPostgressStore() (*PostgresStore, error) {
	connectionString := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStore{
		db: db,
	}, nil

}

func (storage *PostgresStore) Init() error {
	return storage.createAccountTable()
}

func (storage *PostgresStore) createAccountTable() error {
	query := `create table if not exists account (
		id serial primary key,
		first_name varchar(50),
		last_name varchar(50),
		number serial,
		password_encrypted varchar(100),
		balance serial,
		created_at timestamp
	)`

	_, err := storage.db.Exec(query)
	return err
}

func (store *PostgresStore) CreateAccount(account *Account) error {
	query := `insert into account 
	(first_name, last_name, number, password_encrypted, balance, created_at)
	values ($1, $2, $3, $4, $5, $6)`
	_, err := store.db.Query(
		query,
		account.FirstName,
		account.LastName,
		account.Number,
		account.PasswordEncrypted,
		account.Balance,
		account.CreatedAt)

	if err != nil {
		return err
	}

	// fmt.Printf("%+v\n", response)
	return nil
}

func (store *PostgresStore) UpdateAccount(*Account) error {
	return nil
}

func (store *PostgresStore) DeleteAccount(id int) error {
	_, err := store.db.Query("delete from account where id = $1", id)
	return err
}

func (store *PostgresStore) GetAccounts() ([]*Account, error) {
	rows, err := store.db.Query("select * from account")
	if err != nil {
		return nil, err
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := scanIntoAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (store *PostgresStore) GetAccountByID(id int) (*Account, error) {
	rows, err := store.db.Query("select * from account where id = $1", id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("Account %d não encontrada", id)
}
func (store *PostgresStore) GetAccountByNumber(number int) (*Account, error) {
	rows, err := store.db.Query("select * from account where number = $1", number)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("Account com número %d não encontrada", number)

}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	account := new(Account)
	err := rows.Scan(&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.Number,
		&account.PasswordEncrypted,
		&account.Balance,
		&account.CreatedAt)

	return account, err
}
