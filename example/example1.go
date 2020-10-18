package main

import (
	"fmt"
	"github.com/everbslab/sberbank_id"
	"log"
)

var SbidClientId = "012345670123abcd0123012345678901"
var SbidClientSecret = "QWERTY"

func main() {
	sbc := sberbank_id.New(SbidClientId, SbidClientSecret, &sberbank_id.Config{
		Scope:       "openid name snils gender mobile inn maindoc birthdate verified",
		RedirectUrl: "http://127.0.0.1:8080/login",
	})

	// step 1
	fmt.Println("---step 1-- click on sber id emulator and getting redirect location")
	authcode, err := sbc.AuthRequest()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("**** auth code = ", authcode)

	// step 2
	fmt.Println("---step 2-- get token")
	token, err := sbc.GetToken(authcode)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("**** token = ", authcode)

	// step 3
	fmt.Println("---step 3-- personal data")
	pdata, err := sbc.GetPersonalData(token)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("**** Personal data: %v", pdata)
}
