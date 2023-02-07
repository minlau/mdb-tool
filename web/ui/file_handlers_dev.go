//go:build dev

package ui

import (
	"net/http"
)

func GetStaticDir() http.FileSystem {
	return http.Dir("./web/ui/static")
}
