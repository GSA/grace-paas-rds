// https://golangbyexample.com/generate-random-password-golang/
package main

import (
	"fmt"
	"math/rand"
	"strings"
)

func generatePassword() string {
	fmt.Println("Generating password")
	const min = 2             // minimum number of each type of character
	const passwordLength = 20 // length of password
	var lowerCharSet = "abcdedfghijklmnopqrstuvwxyz"
	var upperCharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var specialCharSet = "!@#$%&*"
	var numberSet = "0123456789"
	var allCharSet = lowerCharSet + upperCharSet + specialCharSet + numberSet
	var password strings.Builder

	// Set special character
	for i := 0; i < min; i++ {
		random := rand.Intn(len(specialCharSet))
		password.WriteString(string(specialCharSet[random]))
	}

	// Set numeric
	for i := 0; i < min; i++ {
		random := rand.Intn(len(numberSet))
		password.WriteString(string(numberSet[random]))
	}

	// Set uppercase
	for i := 0; i < min; i++ {
		random := rand.Intn(len(upperCharSet))
		password.WriteString(string(upperCharSet[random]))
	}

	// Set lowercase
	for i := 0; i < min; i++ {
		random := rand.Intn(len(lowerCharSet))
		password.WriteString(string(lowerCharSet[random]))
	}

	remainingLength := passwordLength - 4*min
	for i := 0; i < remainingLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}
