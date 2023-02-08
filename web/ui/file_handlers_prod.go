//go:build !dev

package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

// https://blog.carlmjohnson.net/post/2021/how-to-use-go-embed/
//
//nolint:typecheck
//go:embed static
var static embed.FS

func GetStaticDir() http.FileSystem {
	dir, err := fs.Sub(static, "static")
	if err != nil {
		panic("failed to embed 'assets' folder. Error: " + err.Error())
	}
	return http.FS(dir)
}
