package main

import (
	"regexp"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	p := generatePassword()

	// Check password is 20 characters long
	if len(p) != 20 {
		t.Errorf("generatePassword() failed. Password is not 20 characters long: Len: %d", len(p))
	}

	// Check password has at least two uppercase letters
	re := regexp.MustCompile(`[A-Z]`)
	m := re.FindAllString(p, -1)
	if len(m) < 2 {
		t.Errorf("generatePassword() failed. Password does not contain at least two uppercase letters: %s %v", p, m)
	}

	// Check password has at least two lowercase letters
	re = regexp.MustCompile(`[a-z]`)
	m = re.FindAllString(p, -1)
	if len(m) < 2 {
		t.Errorf("generatePassword() failed. Password does not contain at least two lowercase letters: %s %v", p, m)
	}

	// Check password has at least two numbers
	re = regexp.MustCompile(`[0-9]`)
	m = re.FindAllString(p, -1)
	if len(m) < 2 {
		t.Errorf("generatePassword() failed. Password does not contain at least two numbers: %s %v", p, m)
	}

	// Check password has at least two special characters
	re = regexp.MustCompile(`[\!\@\#\$\%\&\*]`)
	m = re.FindAllString(p, -1)
	if len(m) < 2 {
		t.Errorf("generatePassword() failed. Password does not contain at least two special characters: %s %v", p, m)
	}
}
