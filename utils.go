package sberbank_id

import (
	"math/rand"
	"net/url"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func buildUrl(portions map[string]string) string {
	uv := url.Values{}

	for i, v := range portions {
		uv.Add(i, v)
	}

	return uv.Encode()
}

func generateRandomRqUID(l int) string {
	res := make([]byte, l)

	for i := range res {
		res[i] = rqUIDcharset[rand.Intn(len(rqUIDcharset))]
	}

	return string(res)
}

func parseUrl(u, key string) (string, error) {
	values, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	params, _ := url.ParseQuery(values.RawQuery)

	return params.Get(key), nil
}

func generateRandomString(slen int) string {
	const charset = "abcdefghiklmnoprstxyzABCDEFGHIKLMNOPRSTXYZ0123456789_-"

	b := make([]byte, slen)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func generateStateHash() string {
	return generateRandomString(8)
}

func generateNonce() string {
	return generateRandomString(16)
}
