package sberbankid

import (
	"math/rand"
	"net/url"
	"time"
)

const (
	rqUIDcharset = "abcdefABCDEF0123456789"
	charsetFull  = "abcdefghiklmnoprstxyzABCDEFGHIKLMNOPRSTXYZ0123456789_-"
)

func buildURL(portions map[string]string) string {
	uv := url.Values{}

	for i, v := range portions {
		uv.Add(i, v)
	}

	return uv.Encode()
}

func generateRandomRqUID(l int) string {
	rand.Seed(time.Now().UnixNano())

	res := make([]byte, l)

	for i := range res {
		res[i] = rqUIDcharset[rand.Intn(len(rqUIDcharset))] // #nosec
	}

	return string(res)
}

func parseURL(u, key string) (string, error) {
	values, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	params, _ := url.ParseQuery(values.RawQuery)

	return params.Get(key), nil
}

func generateRandomString(slen int) string {
	rand.Seed(time.Now().UnixNano())

	b := make([]byte, slen)
	for i := range b {
		b[i] = charsetFull[rand.Intn(len(charsetFull))] // #nosec
	}

	return string(b)
}

func isURL(str string) bool {
	u, err := url.Parse(str)

	return err == nil && u.Scheme != "" && u.Host != ""
}
