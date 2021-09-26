package control

import (
	"io/fs"
	"os"
)

func OSStat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}
