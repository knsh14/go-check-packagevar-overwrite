package main

import (
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"

	"github.com/knsh14/go-check-packagevar-overwrite/checker"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("arg is not 1")
		return
	}

	pkg := filepath.Join(build.Default.GOPATH, "src", flag.Arg(0))
	_, err := os.Stat(pkg)
	if err != nil {
		log.Fatal(err)
	}
	err = filepath.Walk(pkg, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			return nil
		}
		pkg, err := build.ImportDir(path, build.ImportComment|build.IgnoreVendor)
		if err != nil {
			if _, ok := err.(*build.NoGoError); ok {
				return nil
			}
			return err
		}
		msgs, err := checker.CheckPkg(pkg)
		if err != nil {
			return err
		}
		if len(msgs) > 0 {
			for _, msg := range msgs {
				for _, t := range msg.Texts {
					fmt.Printf("%s:%d %s\n", msg.Path, msg.Line, t)
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
