package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pin := "0000"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	fmt.Println(string(hash))
}
