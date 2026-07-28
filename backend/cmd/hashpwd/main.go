package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pwd := "password"
	if len(os.Args) > 1 {
		pwd = os.Args[1]
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(hash))
}
