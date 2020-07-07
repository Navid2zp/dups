package dups

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"github.com/cespare/xxhash"
	"github.com/cheggaaa/pb/v3"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

const (
	XXHash = "xxhash"
	MD5    = "md5"
	SHA256 = "sha256"
)

type FileInfo struct {
	Path string
	Info os.FileInfo
}

func getXXHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()
	h := xxhash.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum64()), nil
}

func getMD5(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func GetFileHash(path, algorithm string) (string, error) {
	switch algorithm {
	case MD5:
		return getMD5(path)
	case XXHash:
		return getXXHash(path)
	case SHA256:
		return getSHA256(path)
	default:
		return getMD5(path)
	}
}

func GetFiles(root string, full bool) ([]FileInfo, error) {
	var filesInfos []FileInfo
	cleanedPath := CleanPath(root)
	if full {
		files, err := ioutil.ReadDir(cleanedPath)
		if err != nil {
			return filesInfos, err
		}
		for _, file := range files {
			if !file.IsDir() {
				filesInfos = append(filesInfos, FileInfo{
					Path: filepath.Join(cleanedPath, file.Name()),
					Info: file,
				})
			}
		}
	} else {
		err := filepath.Walk(cleanedPath, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				filesInfos = append(filesInfos, FileInfo{
					Path: path,
					Info: info,
				})
			}
			return nil
		})
		if err != nil {
			return filesInfos, err
		}
	}
	return filesInfos, nil
}

func CollectHashes(files []FileInfo, singleThread bool, minSize int, algorithm string, flat bool) map[string][]FileInfo {
	hashes := map[string][]FileInfo{}
	var lock = sync.RWMutex{}

	var readHash = func(key string) []FileInfo {
		lock.RLock()
		defer lock.RUnlock()
		return hashes[key]
	}

	var writeHash = func(hash string, files []FileInfo) {
		lock.Lock()
		defer lock.Unlock()
		hashes[hash] = files
	}
	var bar *pb.ProgressBar
	if !flat {
		bar = createBar(len(files))
	}

	if singleThread {
		for _, file := range files {
			if int(file.Info.Size()) >= minSize {
				hash, err := GetFileHash(file.Path, algorithm)
				if err == nil {
					hashes[hash] = append(hashes[hash], file)
				}
			}
			if bar != nil {
				bar.Increment()
			}
		}
		if bar != nil {
			bar.Finish()
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(len(files))
		for _, file := range files {
			go func(f FileInfo, bar *pb.ProgressBar) {
				if int(file.Info.Size()) >= minSize {
					hash, err := GetFileHash(f.Path, algorithm)
					if err == nil {
						oldHashes := readHash(hash)
						newHashes := append(oldHashes, f)
						writeHash(hash, newHashes)
					}
				}
				if bar != nil {
					bar.Increment()
				}
				wg.Done()
			}(file, bar)
		}
		wg.Wait()
		if bar != nil {
			bar.Finish()
		}
	}
	return hashes
}

func GetDuplicates(hashes map[string][]FileInfo) ([][]FileInfo, int, int) {
	var duplicateFiles [][]FileInfo
	// total duplicate files
	total := 0
	// Total number of files with duplicates
	totalFiles := 0
	for _, files := range hashes {
		if len(files) > 1 {
			totalFiles++
			// for original file which will be counted in the next for
			total--
			var duplicates []FileInfo
			for _, file := range files {
				total++
				duplicates = append(duplicates, file)
			}
			duplicateFiles = append(duplicateFiles, duplicates)
		}
	}
	return duplicateFiles, totalFiles, total
}

func RemoveDuplicates(fileSets [][]FileInfo) (int, int, error) {
	totalSize := 0
	totalDeleted := 0
	for _, files := range fileSets {
		for i, file := range files {
			if i > 0 {
				totalSize += int(file.Info.Size())
				totalDeleted++
				err := os.Remove(file.Path)
				if err != nil {
					return totalSize, totalDeleted, err
				}

			}
		}
	}
	return totalSize, totalDeleted, nil
}
