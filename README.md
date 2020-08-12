<p align="center">
	<img alt="dups" src="https://raw.githubusercontent.com/Navid2zp/dups/master/dups-h.png" />
</p>


# dups
dups is a CLI tool to find and remove duplicate files using different hash algorithms (MD5, SHA256 and XXHash) with multi-core support.

### Install
Download binaries:

[Release Page][1]

To use in a go project:
```
go get github.com/Navid2zp/dups
```

### Usage

#### CLI

Available Commands:

| Command | Description | 
|---|---|
| clean | Finds duplicate files in a given path and deletes them.  |
| scan |  Finds duplicate files in a given path but doesn't delete them. | 
| help |  Help about any command |

Flags:

| flag | Description | 
|---|---|
| --algorithm string | algorithm to use (md5/sha256/xxhash) (default "md5")  |
| -f, --flat |  flat output, no extra info (only prints duplicate files) | 
| -r, --full |  full search (search in sub-directories too) |
| --min-size int | minimum file size to scan in bytes (default 10) |
| -s, --single-core | use single cpu core |


**Examples:**

Remove duplicates bigger than 1KB using multiple cpu cores:
```
dups clean path/to/directory --min-size 1024
```

Find duplicates and write them into `file.txt`:
```
dups scan path/to/directory -f >> file.txt
```

Find and list duplicates using single cpu core and `XXHash` algorithm:
```
dups scan path/to/directory -s --algorithm xxhash
```

#### Go code:

```go
package main

import (
	"fmt"
	"github.com/Navid2zp/dups"
)

func main()  {
	// list all files including files in any sub-directory
	files, err := dups.GetFiles("path/to/directory", true)
	if err != nil {
		panic(err)
	}
	// collect hashes for all files
	// singleThread: use a single thread
	// flatt: don't print the process bar or any other information
	hashes := dups.CollectHashes(files, false, 1024, dups.XXHash, false)
	duplicates, filesCount, duplicatesCount := dups.GetDuplicates(hashes)
	fmt.Println("total of files with duplicates:", filesCount)
	fmt.Println("total of duplicate files:", duplicatesCount)

	freedSize, deletedCount, err := dups.RemoveDuplicates(duplicates)
	if err != nil {
		panic(err)
	}
	fmt.Println("remove", deletedCount, "files")
	fmt.Println("freed a total of ", freedSize, "bytes")
}
```

#### Notes:

- Use single core option (`-s`) if files are big (depending on your disk type).
- Use XXHash algorithm for fast scanning and MD5/SHA256 for safest scanning or if the number of files is huge.

#### Build from source:

`go build -tags multicore` if you are building using Go < 1.5 or edit `runtime.GOMAXPROCS()` manually to support multi-core.


License
----

[Apache][2]


[1]: https://github.com/Navid2zp/dups/releases
[2]: https://github.com/Navid2zp/dups/blob/master/LICENSE
