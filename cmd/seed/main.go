package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	fmt.Println("hash:", string(hash))

	// 驗證是否能比對回來
	err = bcrypt.CompareHashAndPassword(hash,
		[]byte(password))
	fmt.Println("compare error:", err) // 應該<nil>
}
