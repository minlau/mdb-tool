//go:build !dev

package main

import (
	"embed"
	"io/fs"
	"net/http"
)

// https://blog.carlmjohnson.net/post/2021/how-to-use-go-embed/
//
//go:embed static
var static embed.FS

func getStaticDir() http.FileSystem {
	dir, err := fs.Sub(static, "static")
	if err != nil {
		panic("failed to embed 'assets' folder. Error: " + err.Error())
	}
	return http.FS(dir)
}
