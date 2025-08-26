package service

import "fmt"

func SendOTPEmail(email, otp string) error {
	
	fmt.Printf("Sending OTP %s to email %s\n", otp, email)
	return nil
}
