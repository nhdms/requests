package requests

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gospider007/gson"
)

func readCookies(h http.Header, filter string) []*http.Cookie {
	lines := h["Cookie"]
	if len(lines) == 0 {
		return []*http.Cookie{}
	}

	var cookies []*http.Cookie
	for _, line := range lines {
		parts := strings.Split(strings.TrimSpace(line), ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			name, val, _ := strings.Cut(part, "=")
			name = strings.TrimSpace(name)
			val = strings.TrimSpace(val)
			if name != "" {
				cookies = append(cookies, &http.Cookie{Name: name, Value: val})
			}
		}
	}
	return cookies
}

func readSetCookies(h http.Header) []*http.Cookie {
	lines := h["Set-Cookie"]
	if len(lines) == 0 {
		return []*http.Cookie{}
	}

	var cookies []*http.Cookie
	for _, line := range lines {
		parts := strings.Split(line, ";")
		if len(parts) == 0 {
			continue
		}

		// Parse the name=value part
		nameVal := strings.TrimSpace(parts[0])
		name, val, found := strings.Cut(nameVal, "=")
		if !found {
			continue
		}

		cookie := &http.Cookie{
			Name:  strings.TrimSpace(name),
			Value: strings.TrimSpace(val),
		}

		// Parse attributes
		for _, part := range parts[1:] {
			attr := strings.TrimSpace(part)
			if attr == "" {
				continue
			}

			key, value, hasValue := strings.Cut(attr, "=")
			key = strings.ToLower(strings.TrimSpace(key))

			switch key {
			case "domain":
				if hasValue {
					cookie.Domain = strings.TrimSpace(value)
				}
			case "path":
				if hasValue {
					cookie.Path = strings.TrimSpace(value)
				}
			case "secure":
				cookie.Secure = true
			case "httponly":
				cookie.HttpOnly = true
			}
		}

		cookies = append(cookies, cookie)
	}
	return cookies
}

// 支持json,map,[]string,http.Header,string
func ReadCookies(val any) (Cookies, error) {
	switch cook := val.(type) {
	case *http.Cookie:
		return Cookies{
			cook,
		}, nil
	case http.Cookie:
		return Cookies{
			&cook,
		}, nil
	case Cookies:
		return cook, nil
	case []*http.Cookie:
		return Cookies(cook), nil
	case string:
		return readCookies(http.Header{"Cookie": []string{cook}}, ""), nil
	case http.Header:
		return readCookies(cook, ""), nil
	case []string:
		return readCookies(http.Header{"Cookie": cook}, ""), nil
	default:
		return any2Cookies(cook)
	}
}

func ReadSetCookies(val any) (Cookies, error) {
	switch cook := val.(type) {
	case Cookies:
		return cook, nil
	case []*http.Cookie:
		return Cookies(cook), nil
	case string:
		return readSetCookies(http.Header{"Set-Cookie": []string{cook}}), nil
	case http.Header:
		return readSetCookies(cook), nil
	case []string:
		return readSetCookies(http.Header{"Set-Cookie": cook}), nil
	default:
		return any2Cookies(cook)
	}
}
func any2Cookies(val any) (Cookies, error) {
	switch cooks := val.(type) {
	case map[string]string:
		cookies := Cookies{}
		for kk, vv := range cooks {
			cookies = append(cookies, &http.Cookie{
				Name:  kk,
				Value: vv,
			})
		}
		return cookies, nil
	case map[string][]string:
		cookies := Cookies{}
		for kk, vvs := range cooks {
			for _, vv := range vvs {
				cookies = append(cookies, &http.Cookie{
					Name:  kk,
					Value: vv,
				})
			}
		}
		return cookies, nil
	case *gson.Client:
		if !cooks.IsObject() {
			return nil, errors.New("cookies not support type")
		}
		cookies := Cookies{}
		for kk, vvs := range cooks.Map() {
			if vvs.IsArray() {
				for _, vv := range vvs.Array() {
					cookies = append(cookies, &http.Cookie{
						Name:  kk,
						Value: vv.String(),
					})
				}
			} else {
				cookies = append(cookies, &http.Cookie{
					Name:  kk,
					Value: vvs.String(),
				})
			}
		}
		return cookies, nil
	default:
		jsonData, err := gson.Decode(cooks)
		if err != nil {
			return nil, err
		}
		cookies := Cookies{}
		for kk, vvs := range jsonData.Map() {
			if vvs.IsArray() {
				for _, vv := range vvs.Array() {
					cookies = append(cookies, &http.Cookie{
						Name:  kk,
						Value: vv.String(),
					})
				}
			} else {
				cookies = append(cookies, &http.Cookie{
					Name:  kk,
					Value: vvs.String(),
				})
			}
		}
		return cookies, nil
	}
}
func (obj *RequestOption) initCookies() (err error) {
	if obj.Cookies == nil {
		return nil
	}
	obj.Cookies, err = ReadCookies(obj.Cookies)
	return err
}
