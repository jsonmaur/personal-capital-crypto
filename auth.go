package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/go-redis/redis/v8"
	"golang.org/x/net/publicsuffix"
)

type ApiResponse struct {
	SpHeader SpHeader               `json:"spHeader"`
	SpData   map[string]interface{} `json:"spData"`
}

type SpHeader struct {
	CSRF       string  `json:"csrf"`
	Success    bool    `json:"success"`
	Status     string  `json:"status"`
	AuthLevel  string  `json:"authLevel"`
	UserGuid   string  `json:"userGuid"`
	PersonId   int     `json:"personId"`
	Username   string  `json:"username"`
	DeviceName string  `json:"deviceName"`
	Errors     []Error `json:"errors"`
}

type Error struct {
	Message string `json:"message"`
}

const (
	API_BASE_URL      = "https://home.personalcapital.com"
	SESSION_REDIS_KEY = "personalcapital:session"

	AUTH_LEVEL_IDENTIFIED     = "USER_IDENTIFIED"
	AUTH_LEVEL_REMEMBERED     = "USER_REMEMBERED"
	AUTH_LEVEL_AUTHORIZED     = "DEVICE_AUTHORIZED"
	AUTH_LEVEL_AUTHENTTICATED = "SESSION_AUTHENTICATED"
)

var client *http.Client
var csrf string

func requestAPI(path string, data url.Values) *ApiResponse {
	uri := fmt.Sprintf("%s/api/%s", API_BASE_URL, path)
	body := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", uri, body)
	check(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	var res ApiResponse
	json.NewDecoder(resp.Body).Decode(&res)

	for _, v := range res.SpHeader.Errors {
		log.Printf("Error: %s", v)
	}

	return &res
}

func initialCSRF() string {
	uri := fmt.Sprintf("%s/page/login/goHome", API_BASE_URL)
	req, err := http.NewRequest("GET", uri, nil)
	check(err)

	resp, err := client.Do(req)
	check(err)
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	pattern := regexp.MustCompile("(?:globals.csrf=)|(?:csrf = )'([a-zA-Z0-9-]+)'")

	for scanner.Scan() {
		v := pattern.FindStringSubmatch(scanner.Text())
		if len(v) > 0 {
			return v[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return ""
}

func isAuthenticated(csrf string) bool {
	res := requestAPI("login/querySession", url.Values{
		"csrf":               {csrf},
		"apiClient":          {"WEB"},
		"lastServerChangeId": {"-1"},
	})

	return res.SpHeader.AuthLevel == AUTH_LEVEL_AUTHENTTICATED
}

func persistCookies(jar *cookiejar.Jar) {
	base, err := url.Parse(API_BASE_URL)
	check(err)

	marshaled, err := json.Marshal(jar.Cookies(base))
	check(err)

	err = db.Set(ctx, SESSION_REDIS_KEY, marshaled, 0).Err()
	check(err)
}

func restoreCookies(jar *cookiejar.Jar) {
	base, err := url.Parse(API_BASE_URL)
	check(err)

	data, err := db.Get(ctx, SESSION_REDIS_KEY).Result()
	if err == redis.Nil {
		return
	}
	check(err)

	var cookies []*http.Cookie
	err = json.Unmarshal([]byte(data), &cookies)
	check(err)

	jar.SetCookies(base, cookies)
}

func authenticateUser(csrf, username string) *ApiResponse {
	res := requestAPI("login/identifyUser", url.Values{
		"csrf":            {csrf},
		"apiClient":       {"WEB"},
		"bindDevice":      {"false"},
		"username":        {username},
		"skipLinkAccount": {"false"},
		"skipFirstUse":    {},
		"redirectTo":      {},
		"referrerId":      {},
	})

	if res.SpHeader.Status != "ACTIVE" {
		log.Fatal("User account is not active")
	}

	return res
}

func authenticateMFA(csrf string) *ApiResponse {
	method := "authenticateSms"
	data := url.Values{
		"csrf":            {csrf},
		"apiClient":       {"WEB"},
		"bindDevice":      {"false"},
		"challengeMethod": {"OP"},
		"challengeReason": {"DEVICE_AUTH"},
	}

	res := requestAPI("credential/challengeSms", data)
	if !res.SpHeader.Success {
		method = "authenticateEmailByCode"
		res = requestAPI("credential/challengeEmail", data)
		if !res.SpHeader.Success {
			method = "authenticatePhone"
			res = requestAPI("credential/challengePhone", data)
			if !res.SpHeader.Success {
				log.Fatal("Multi-Factor Authentication failed")
			}
		}
	}

	fmt.Print("Multi-Factor Authentication Code: ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()

	res = requestAPI(fmt.Sprintf("credential/%s", method), url.Values{
		"csrf":            {csrf},
		"apiClient":       {"WEB"},
		"bindDevice":      {"false"},
		"challengeMethod": {"OP"},
		"challengeReason": {"DEVICE_AUTH"},
		"code":            {input.Text()},
	})

	if res.SpHeader.AuthLevel != AUTH_LEVEL_AUTHORIZED {
		log.Fatal("Multi-Factor Authentication failed")
	}

	return res
}

func authenticatePass(csrf, username, password string) *ApiResponse {
	res := requestAPI("credential/authenticatePassword", url.Values{
		"csrf":            {csrf},
		"apiClient":       {"WEB"},
		"bindDevice":      {"true"},
		"deviceName":      {"Personal Capital Crypto"},
		"username":        {username},
		"passwd":          {password},
		"skipLinkAccount": {"false"},
		"skipFirstUse":    {},
		"redirectTo":      {},
		"referrerId":      {},
	})

	return res
}

func Authenticate() string {
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	check(err)
	restoreCookies(jar)
	defer persistCookies(jar)

	client = &http.Client{Jar: jar}
	csrf := initialCSRF()

	if !isAuthenticated(csrf) {
		user := authenticateUser(csrf, CFG_PC_USERNAME)

		switch user.SpHeader.AuthLevel {
		case AUTH_LEVEL_IDENTIFIED:
			authenticateMFA(user.SpHeader.CSRF)
			authenticatePass(user.SpHeader.CSRF, CFG_PC_USERNAME, CFG_PC_PASSWORD)
		case AUTH_LEVEL_REMEMBERED:
			authenticatePass(user.SpHeader.CSRF, CFG_PC_USERNAME, CFG_PC_PASSWORD)
		}

		return user.SpHeader.CSRF
	}

	return csrf
}
