# Sberbank ID

Sberbank ID client implementation.

## Usage
Install module
```Bash
go get github.com/everbslab/sberbankid
```

Apply in your app:
```Go
    var (
        SbidClientId = "0123456.....12345678901"    // yours client id
        SbidClientSecret = "Q....Y"                 // your client secret
    )

    sbc := sberbankid.New(SbidClientId, SbidClientSecret, &sberbank_id.Config{
        Scope:       "openid name snils gender mobile inn maindoc birthdate verified",
        RedirectUrl: "http://127.0.0.1:8080/login",
        Env:         sberbankid.EnvSandbox,
        VerboseMode: true,
    })

    authcode, err := sbc.AuthRequest("Q0002", "Password2")
    if err != nil {
        log.Fatal(err)
    }

    token, err := sbc.GetToken(authcode)
    if err != nil {
        log.Fatal(err)
    }

    if pdata, err := sbc.GetPersonalData(token); err == nil {
        log.Printf("**** Personal data map: %v\n", pdata)
    } else {
        log.Fatal(err)
    }
```

Result will be data map with Personal Data received from Sberbank according to defined `scope`.
```Bash
{"iss":"http://45.12.238.224:8181/ru/prod/tokens/v2/oidc","sub":"022fa8480ff243439f5887ab5a847c1b","aud":"012345670123abcd0123012345678901","birthdate":"1980.01.01","identificaton":{"series":"0001","number":"000001","issuedBy":null,
"issuedDate":"2000.01.01","code":"001-001"},"inn":{"number":"0000000001"},"snils":{"number":"0000001"},"gender":1,"verified":1,"family_name":"Иванов","given_name":"Иван","middle_name":"Иванович","phone_number":"+79001000001"}
2020/10/19 15:55:37 **** Personal data: &map[aud:012345670123abcd0123012345678901 birthdate:1980.01.01 family_name:Иванов gender:1 given_name:Иван identificaton:map[code:001-001 issuedBy:<nil> issuedDate:2000.01.01 number:000001 ser
ies:0001] inn:map[number:0000000001] iss:http://45.12.238.224:8181/ru/prod/tokens/v2/oidc middle_name:Иванович phone_number:+79001000001 snils:map[number:0000001] sub:022fa8480ff243439f5887ab5a847c1b verified:1]
```


Protocol specifications:
* [Tech spec](https://developer.sberbank.ru/doc/v1/sberbank-id/info)
* [Sandbox](https://developer.sberbank.ru/doc/v1/sberbank-id/Sand)