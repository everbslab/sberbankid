package sberbankid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	errWrongEnvironment    = errors.New("wrong environment config")
	errNotValidURL         = errors.New("not valid URL")
	errAuthRequest         = errors.New("auth request failed")
	errPersonalDataRequest = errors.New("failed to fetch personal data")
)

const (
	rqUIDcharset  = "abcdefABCDEF0123456789"
	rqUIDsize     = 32 // RqUID constant length
	stateHashSize = 8  // randomly generated state hash length
	nonceHashSize = 16 // randomly generated nonce hash length

	// API endpoints URIs.
	AuthorizeAccessTokenURI = "/CSAFront/oidc/sberbank_id/authorize.do" // #nosec
	TokenAuthorizeURI       = "/ru/prod/tokens/v2/oidc"                 // #nosec
	PersonalDataURI         = "/ru/prod/sberbankid/v2.1/userInfo"       // #nosec

	// Endpoints for environments: urls.
	EndpointDev     = "https://dev.api.sberbank.ru/"
	EndpointProd    = "https://sec.api.sberbank.ru"
	EndpointSandbox = "http://45.12.238.224:8181"

	// Environment options.
	EnvSandbox Environment = 1 << iota
	EnvDev
	EnvProd
)

// Client represents the main struct of a client for Sbebank ID API.
type Client struct {
	HTTPCient *http.Client
	creds     *SberCredentials
	config    *Config
}

// SberCredentials is a struct to store user credentials for Sberbank ID API.
type SberCredentials struct {
	ClientID     string
	ClientSecret string
}

// TokenResponse represents OAuth token response.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	Scope       string `json:"scope"`
	IDToken     string `json:"id_token"`
}

// PersonData represents Personal Data response. JSON formatted.
type PersonData map[string]interface{}

// Environment type.
type Environment int

// Config is a structure to keep instance parameters for httpClient requests.
type Config struct {
	Scope       string
	RedirectURL string
	state       string
	nonce       string
	Env         Environment
	VerboseMode bool
}

// NewClient initializes Client instance.
func NewClient(clientID, clientSecret string, config *Config) *Client {
	if config.Env == 0 {
		config.Env = EnvSandbox // default target environment
	}

	return &Client{
		HTTPCient: &http.Client{
			// prevents redirects follow
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		config: &Config{
			Scope:       config.Scope,
			RedirectURL: config.RedirectURL,
			state:       generateStateHash(stateHashSize),
			nonce:       generateNonce(nonceHashSize),
			Env:         config.Env,
			VerboseMode: config.VerboseMode,
		},
		creds: &SberCredentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		},
	}
}

func (c *Client) GetToken(authcode string) (*TokenResponse, error) {
	rm := map[string]string{
		"grant_type":    "authorization_code",
		"scope":         c.config.Scope,
		"redirect_uri":  c.config.RedirectURL,
		"code":          authcode,
		"client_id":     c.creds.ClientID,
		"client_secret": c.creds.ClientSecret,
	}

	url, err := c.GetEnvURL(TokenAuthorizeURI)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(buildURL(rm)))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("X-IBM-Client-ID", c.creds.ClientID)
	req.Header.Add("X-IBM-Client-Secret", c.creds.ClientSecret)
	req.Header.Add("RqUID", generateRandomRqUID(rqUIDsize))

	resp, err := c.HTTPCient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if c.config.VerboseMode {
		fmt.Println("--------------------")
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		fmt.Println("response Body:", string(body))
	}

	tr := &TokenResponse{}
	if err := json.Unmarshal(body, &tr); err != nil {
		return nil, err
	}

	return tr, nil
}

func (c *Client) AuthRequest(login, pass string) (string, error) {
	userCredMap := map[string]string{
		"login":    login,
		"password": pass,
	}

	queryMap := map[string]string{
		"response_type": "code",
		"client_type":   "PRIVATE",
		"scope":         c.config.Scope,
		"client_id":     c.creds.ClientID,
		"state":         c.config.state,
		"nonce":         c.config.nonce,
		"redirect_uri":  c.config.RedirectURL,
	}

	jsonCreds, err := json.Marshal(userCredMap)
	if err != nil {
		return "", err
	}

	url, err := c.GetEnvURL(AuthorizeAccessTokenURI + "?" + buildURL(queryMap))
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonCreds))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPCient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if c.config.VerboseMode {
		fmt.Println(string(jsonCreds))
		fmt.Println("---")
		fmt.Println("req", req)
		fmt.Println("--------------------")
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	}

	// We need to catch code 200
	if resp.StatusCode == http.StatusOK {
		loc := resp.Header.Get("Location")
		if loc != "" {
			authCode, _ := parseURL(resp.Header.Get("Location"), "code")

			return authCode, nil
		}
	}

	return "", errAuthRequest
}

func (c *Client) GetPersonalData(token *TokenResponse) (*PersonData, error) {
	url, err := c.GetEnvURL(PersonalDataURI)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-introspect-rquid", generateRandomRqUID(rqUIDsize))
	req.Header.Add("X-IBM-Client-ID", c.creds.ClientID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token.AccessToken))
	req.Header.Add("Accept", "application/json")
	req.Header.Add("RqUID", generateRandomRqUID(rqUIDsize))

	res, err := c.HTTPCient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)

	if c.config.VerboseMode {
		fmt.Println(string(body))
	}

	if err != nil {
		return nil, err
	}

	var pd PersonData
	if err := json.Unmarshal(body, &pd); err != nil {
		return nil, errPersonalDataRequest
	}

	return &pd, nil
}

func (c *Client) GetEnvURL(uri string) (string, error) {
	envMap := map[Environment]string{
		EnvDev:     EndpointDev,
		EnvSandbox: EndpointSandbox,
		EnvProd:    EndpointProd,
	}

	if endpoint, ok := envMap[c.config.Env]; ok {
		url := endpoint + uri
		if isURL(url) {
			return url, nil
		}

		return "", errNotValidURL
	}

	return "", errWrongEnvironment
}
