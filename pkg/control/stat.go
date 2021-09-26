package control

import (
	"io/fs"
	"os"
)

func StatOS(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}
