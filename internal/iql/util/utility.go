package util

import (
	"path"
	"path/filepath"
	"runtime"
)

func GetFilePathFromRepositoryRoot(relativePath string) (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	curDir := filepath.Dir(filename)
	return filepath.Abs(path.Join(curDir, "../../..", relativePath))
}

func MaxMapKey(numbers map[int]interface{}) int {
	var maxNumber int
	for maxNumber = range numbers {
		break
	}
	for n := range numbers {
		if n > maxNumber {
			maxNumber = n
		}
	}
	return maxNumber
}
