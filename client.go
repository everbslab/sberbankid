package sberbank_id

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	rqUIDcharset = "abcdefABCDEF0123456789"
	rqUIDlen     = 32

	// refer tot https://developer.sberbank.ru/doc/v1/sberbank-id/tokensurl
	//ApiUrlAuthDev  = "https://dev.api.sberbank.ru/ru/prod/tokens/v2/oidc"
	//ApiUrlAuthProd = "https://sec.api.sberbank.ru/ru/prod/tokens/v2/oidc"

	AuthorizeAccessTokenUrl = "http://45.12.238.224:8181/CSAFront/oidc/sberbank_id/authorize.do"
	TokenAuthorizeUrl       = "http://45.12.238.224:8181/ru/prod/tokens/v2/oidc"
	PersonalDataUrl         = "http://45.12.238.224:8181/ru/prod/sberbankid/v2.1/userInfo"
)

type SberbankIdClient struct {
	HttpCient *http.Client
	creds     *SberCredentials
	config    *Config
}

type SberCredentials struct {
	ClientId     string
	ClientSecret string
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope"`
	IdToken     string `json:"id_token"`
}

// PersonData represents Personal Data response. JSON formatted
type PersonData map[string]interface{}

type Config struct {
	Scope       string
	RedirectUrl string
	state       string
	nonce       string
}

func New(clientId, clientSecret string, config *Config) *SberbankIdClient {
	return &SberbankIdClient{
		HttpCient: &http.Client{
			// prevents redirects follow
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		config: &Config{
			Scope:       config.Scope,
			RedirectUrl: config.RedirectUrl,
			state:       generateStateHash(),
			nonce:       generateNonce(),
		},
		creds: &SberCredentials{
			ClientId:     clientId,
			ClientSecret: clientSecret,
		},
	}
}

func (c *SberbankIdClient) GetToken(authcode string) (*TokenResponse, error) {
	fmt.Println("Getting token...")

	rm := make(map[string]string)
	rm["grant_type"] = "authorization_code"
	rm["scope"] = c.config.Scope
	rm["redirect_uri"] = c.config.RedirectUrl
	rm["code"] = authcode
	rm["client_id"] = c.creds.ClientId
	rm["client_secret"] = c.creds.ClientSecret

	req, _ := http.NewRequest("POST", TokenAuthorizeUrl, bytes.NewBufferString(buildUrl(rm)))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-IBM-Client-ID", c.creds.ClientId)
	req.Header.Add("X-IBM-Client-Secret", c.creds.ClientSecret)
	req.Header.Add("RqUID", generateRandomRqUID(rqUIDlen))

	fmt.Println("--------------------")
	fmt.Println(req)

	resp, err := c.HttpCient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Println("--------------------")
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	tresp := &TokenResponse{}
	if err := json.Unmarshal(body, tresp); err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Println(tresp)

	return tresp, nil
}

func (c *SberbankIdClient) AuthRequest() (string, error) {
	var userCredMap = map[string]string{
		"login":    "Q0002",
		"password": "Password2",
	}

	var qparams = map[string]string{
		"response_type": "code",
		"client_type":   "PRIVATE",
		"scope":         c.config.Scope,
		"client_id":     c.creds.ClientId,
		"state":         c.config.state,
		"nonce":         c.config.nonce,
		"redirect_uri":  c.config.RedirectUrl,
	}

	jsonCreds, err := json.Marshal(userCredMap)
	if err != nil {
		return "", err
	}
	fmt.Println(string(jsonCreds))
	fmt.Println("---")

	req, _ := http.NewRequest("POST", AuthorizeAccessTokenUrl+"?"+buildUrl(qparams), bytes.NewBuffer(jsonCreds))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HttpCient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}

	fmt.Println("req", req)
	fmt.Println("--------------------")
	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	// We need to catch code 200
	if resp.StatusCode == http.StatusOK {
		fmt.Println("PARSE LOCATION")
		loc := resp.Header.Get("Location")
		if loc != "" {
			authCode, _ := parseUrl(resp.Header.Get("Location"), "code")
			return authCode, nil
		}
	}

	return "", errors.New("auth request failed")
}

func (c *SberbankIdClient) GetPersonalData(token *TokenResponse) (*PersonData, error) {
	req, err := http.NewRequest("GET", PersonalDataUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-introspect-rquid", generateRandomRqUID(32))
	req.Header.Add("X-IBM-Client-ID", c.creds.ClientId)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token.AccessToken))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("RqUID", generateRandomRqUID(rqUIDlen))

	res, err := c.HttpCient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	pd := make(PersonData)
	if err := json.Unmarshal(body, &pd); err != nil {
		return nil, err
	}

	return &pd, nil
}
