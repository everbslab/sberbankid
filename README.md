# Sberbank ID

Sberbank ID client implementation.

## Usage
```Go
    var SbidClientId = "0123456.....12345678901"
    var SbidClientSecret = "Q....Y"

    sbc := sberbank_id.New(SbidClientId, SbidClientSecret, &sberbank_id.Config{
        Scope:       "openid name snils gender mobile inn maindoc birthdate verified",
        RedirectUrl: "http://127.0.0.1:8080/login",
        DebugMode:   true,
    })

    authcode, err := sbc.AuthRequest()
    if err != nil {
        log.Fatal(err)
    }

    token, err := sbc.GetToken(authcode)
    if err != nil {
        log.Fatal(err)
    }

    if pdata, err := sbc.GetPersonalData(token); err == nil {
        logInfo(fmt.Sprintf("**** Personal data map: %v", pdata))
    } else {
        log.Fatal(err)
    }
```

Protocol specifications:
* [Tech spec](https://developer.sberbank.ru/doc/v1/sberbank-id/info)
* [Sandbox](https://developer.sberbank.ru/doc/v1/sberbank-id/Sand)