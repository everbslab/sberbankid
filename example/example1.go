package main

import (
	"fmt"
	"log"

	"github.com/everbslab/sberbankid"
)

func main() {
	sbc := sberbankid.NewClient("012345670123abcd0123012345678901", "QWERTY", &sberbankid.Config{
		Scope:       "openid name snils gender mobile inn maindoc birthdate verified",
		RedirectURL: "http://127.0.0.1:8080/login",
		Env:         sberbankid.EnvSandbox,
		VerboseMode: true,
	})

	// step 1
	logInfo("---step 1-- click on sberbank id btn emulator and getting redirect location")
	authcode, err := sbc.AuthRequest("Q0002", "Password2")
	if err != nil {
		log.Fatal(err)
	}
	logInfo(fmt.Sprintf("**** auth code = %v", authcode))

	// step 2
	logInfo("---step 2-- get token")
	token, err := sbc.GetToken(authcode)
	if err != nil {
		log.Fatal(err)
	}
	logInfo(fmt.Sprintf("**** token = %v", token.AccessToken))

	// step 3
	logInfo("---step 3-- personal data")
	if pdata, err := sbc.GetPersonalData(token); err == nil {
		logInfo(fmt.Sprintf("**** Personal data map: %v", pdata))
	} else {
		log.Fatal(err)
	}
}

func logInfo(str string) {
	log.Println(str)
}
