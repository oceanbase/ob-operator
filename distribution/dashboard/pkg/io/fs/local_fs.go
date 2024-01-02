package fs

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gin-contrib/static"
)

const INDEX = "index.html"

type localFileSystem struct {
	http.FileSystem
	root    string
	indexes bool
}

func (l *localFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)
		stats, err := os.Stat(name)
		if err != nil {
			return false
		}
		if stats.IsDir() {
			if !l.indexes {
				index := path.Join(name, INDEX)
				_, err := os.Stat(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

func LocalFSWrapper(fs http.FileSystem) static.ServeFileSystem {
	return &localFileSystem{
		FileSystem: fs,
		root:       "/",
		indexes:    false,
	}

}
