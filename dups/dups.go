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
	// XXHash represents XXHash algorithm
	XXHash = "xxhash"
	// MD5 represents XXHash algorithm
	MD5    = "md5"
	// SHA256 represents XXHash algorithm
	SHA256 = "sha256"
)

// FileInfo represents a file containing os.FileInfo and file path
type FileInfo struct {
	Path string
	Info os.FileInfo
}

// getXXHash return xxhash of a file
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

// getMD5 returns md5 hash of a file
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

// getSHA256 returns sha256 hash of a file
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

// GetFileHash returns given file hash using the provided algorithm
// Default: md5
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

// GetFiles finds and returns all the files in the given path
// It will also returns any file in sub-directories if "full=true"
func GetFiles(root string, full bool) ([]FileInfo, error) {
	var filesInfos []FileInfo
	cleanedPath := CleanPath(root)
	if !full {
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

// GroupFiles groups files based on their file size
// This will help avoid unnecessary hash calculations since files with different file sizes can't be duplicates
func GroupFiles(files []FileInfo, minSize int) (map[int][]FileInfo, int) {
	groups := make(map[int][]FileInfo)
	fileCount := 0
	for _, file := range files {
		size := int(file.Info.Size())
		// Ignore files less than minimum size
		if size > minSize {
			groups[size] = append(groups[size], file)
			fileCount++
		}
	}
	return groups, fileCount
}

// CollectHashes returns hashes for the given group files if there is more than one file with the same size
// A hash will be the key and a list of FileInfo for files that share the hash as the value
// "singleThread=false" will force all the function to use one thread only
// minSize is the minimum file size to scan
// "flat=true" will tell the function not to print out any data other than the path to duplicate files
// algorithm is the algorithm to calculate the hash with
func CollectHashes(fileGroups map[int][]FileInfo, singleThread bool, algorithm string, flat bool, fileCount int) map[string][]FileInfo {
	hashes := map[string][]FileInfo{}

	// You cant't read/write at the same time to a map
	// readHash and writeHash will read/write the given key/value to/from the map
	// they make sure that the map is locked while a read or write is happening
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

	// progress bar to show if "flat=false"
	var bar *pb.ProgressBar
	if !flat {
		bar = createBar(fileCount)
	}

	if singleThread {
		for _, group := range fileGroups {
			// Ignore groups with one file
			if len(group) > 1 {
				for _, file := range group {
					hash, err := GetFileHash(file.Path, algorithm)
					if err == nil {
						hashes[hash] = append(hashes[hash], file)
					}
					if bar != nil {
						bar.Increment()
					}
				}
			} else {
				if bar != nil {
					bar.Increment()
				}
			}
		}
		if bar != nil {
			bar.Finish()
		}
	} else {
		var wg sync.WaitGroup
		wg.Add(fileCount)
		for _, group := range fileGroups {
			// Ignore groups with one file
			if len(group) > 1 {
				for _, file := range group {
					go func(f FileInfo, bar *pb.ProgressBar) {
						hash, err := GetFileHash(f.Path, algorithm)
						if err == nil {
							oldHashes := readHash(hash)
							newHashes := append(oldHashes, f)
							writeHash(hash, newHashes)
						}
						if bar != nil {
							// tell the progress bar that a process is finished
							bar.Increment()
						}
						wg.Done()
					}(file, bar)
				}
			} else {
				wg.Done()
				if bar != nil {
					bar.Increment()
				}
			}
		}
		wg.Wait()
		if bar != nil {
			// tell the progress bar that all the processes are finished
			bar.Finish()
		}
	}
	return hashes
}

// GetDuplicates scans the given map of hashes and finds the one with duplicates
// It will return a slice containing slices with each slice containing paths to duplicate files
// It will also returns the total of duplicate files and the total of files that have duplicates
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

// RemoveDuplicates removes duplicates
// It will keep the first file in a duplicate set and removes any other files in the set
// It will return the sum of deleted file sizes and total number of deleted files
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
