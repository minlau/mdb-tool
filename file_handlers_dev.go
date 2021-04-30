// +build dev

package main

import (
	"net/http"
)

func getStaticDir() http.FileSystem {
	return http.Dir("./static")
}
