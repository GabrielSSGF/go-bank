package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	seed := flag.Bool("seed", false, "seed the db")
	flag.Parse()

	store, err := newPostgressStore()
	if err != nil {
		log.Fatal(err)
	}

	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	if *seed {
		fmt.Println("seeding o db")
		seedAccounts(store)
	}

	server := newAPIServer(":3000", store)
	server.Run()
}

func seedAccounts(store Storage) {
	seedAccount(store, "Gabriel", "Soares", "teste123")
}

func seedAccount(store Storage, firstName, lastName, password string) *Account {
	account, err := NewAccount(firstName, lastName, password)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.CreateAccount(account); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Nova conta: ", account.Number)

	return account
}
