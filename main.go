package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/knsh14/kamata-tool-two/checker"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("arg is not 1")
		return
	}

	pkg := filepath.Join(build.Default.GOPATH, "src", os.Args[1])
	if dir, err := os.Stat(pkg); err == nil && dir.IsDir() {
		files, err := ioutil.ReadDir(pkg)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range files {
			p := filepath.Join(pkg, f.Name())
			err := checker.Check(p)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
