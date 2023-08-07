package utils

import (
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"time"
)

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func FileInfo(path string) []os.FileInfo {
	dir, _ := ioutil.ReadDir(path)
	return dir
}

func CreateDirectory(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func SubtractTime(b, f time.Time) float64 {
	diff := f.Sub(b).Seconds()
	return math.Floor(diff)
}
