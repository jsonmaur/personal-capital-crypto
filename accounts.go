package main

import (
	"net/url"
)

func GetAccounts(csrf string) *ApiResponse {
	res := requestAPI("newaccount/getAccounts2", url.Values{
		"csrf":               {csrf},
		"apiClient":          {"WEB"},
		"lastServerChangeId": {"-1"},
	})

	return res
}
