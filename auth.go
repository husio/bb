package main

import "net/http"

func CurrentUserID(r *http.Request) (uint, bool) {
	// TODO - authenticate using signed cookie
	return 1, true
}
