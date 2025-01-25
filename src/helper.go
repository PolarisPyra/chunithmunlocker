package main

import (
	"os"
	"path/filepath"
)

func countSpecificXMLFiles(dir string, filenames []string) (map[string]int, error) {
	counts := make(map[string]int)
	for _, filename := range filenames {
		counts[filename] = 0
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			for _, filename := range filenames {
				if info.Name() == filename {
					counts[filename]++
				}
			}
		}
		return nil
	})
	return counts, err
}
