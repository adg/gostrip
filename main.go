/*
Copyright 2015 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Command gostrip builds a minimal Go installation,
// providing the bare minimum required to build Go programs.
//
// See 'gostrip -help' for instruction.
package main

// TODO(adg): support cross-compilation

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	repo = flag.String("repo", "https://go.googlesource.com/go", "Repository location")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %v [flags] <destination>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(2)
	}
	dest := flag.Arg(0)

	if _, err := os.Stat(dest); err == nil {
		dief("destination %v already exists; won't overwrite.\n", dest)
	}

	cmd := exec.Command("git", "clone", *repo, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		dief("cloning repo: %v\n", err)
	}

	if runtime.GOOS == "windows" {
		cmd = exec.Command(".\\make.bat")
	} else {
		cmd = exec.Command("./make.bash")
	}
	cmd.Dir = filepath.Join(dest, "src")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		dief("building go: %v\n", err)
	}

	for _, pat := range alwaysRemove {
		pat = strings.Replace(pat, "GOOS", runtime.GOOS, -1)
		pat = strings.Replace(pat, "GOARCH", runtime.GOARCH, -1)

		name := filepath.Join(dest, pat)
		if err := osAwareRemoveAll(name); err != nil {
			dief("removing %v: %v\n", name, err)
		}
	}

	var remove []string
	if err := filepath.Walk(filepath.Join(dest, "src"), func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() && filepath.Base(path) == "testdata" {
			remove = append(remove, path)
			return filepath.SkipDir
		}
		if !fi.IsDir() && strings.HasSuffix(path, "_test.go") {
			remove = append(remove, path)
		}
		return nil
	}); err != nil {
		dief("looking for test data: %v\n", err)
	}
	for _, name := range remove {
		if err := osAwareRemoveAll(name); err != nil {
			dief("removing %v: %v\n", name, err)
		}
	}
}

func osAwareRemoveAll(path string) error {
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		// skip if file not existed
		return nil
	}
	if err != nil {
		return err
	}

	if runtime.GOOS == "windows" {
		mode := fi.Mode()
		cmd := exec.Command("cmd.exe", "/C", "del", "/Q", "/F", "/S", path)

		if mode.IsDir() {
			cmd = exec.Command("cmd.exe", "/C", "rmdir", "/Q", "/S", path)
		}

		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	return os.RemoveAll(path)
}

func dief(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

var alwaysRemove = []string{
	".git",
	".gitattributes",
	".gitignore",
	"CONTRIBUTING.md",
	"VERSION.cache",
	"favicon.ico",
	"robots.txt",

	"api",
	"include",
	"lib",
	"misc",
	"pkg/obj",
	"test",

	"bin/dist",

	"pkg/GOOS_GOARCH/cmd",

	"pkg/tool/GOOS_GOARCH/dist",
	"pkg/tool/GOOS_GOARCH/fix",
	"pkg/tool/GOOS_GOARCH/nm",
	"pkg/tool/GOOS_GOARCH/objdump",
	"pkg/tool/GOOS_GOARCH/yacc",

	"src/cmd/dist",
	"src/cmd/fix",
	"src/cmd/nm",
	"src/cmd/objdump",
	"src/cmd/yacc",

	"src/cmd/5a",
	"src/cmd/5g",
	"src/cmd/5l",
	"src/cmd/6a",
	"src/cmd/6g",
	"src/cmd/6l",
	"src/cmd/8a",
	"src/cmd/8g",
	"src/cmd/8l",
	"src/cmd/9a",
	"src/cmd/9g",
	"src/cmd/9l",
	"src/cmd/cc",
	"src/cmd/gc",

	"src/cmd/link",
	"src/cmd/ld",
	"src/cmd/pack", // ?

	"src/cmd/go",

	"src/all.bash",
	"src/all.bat",
	"src/all.rc",
	"src/androidtest.bash",
	"src/clean.bash",
	"src/clean.bat",
	"src/clean.rc",
	"src/make.Dist",
	"src/make.bash",
	"src/make.bat",
	"src/make.rc",
	"src/nacltest.bash",
	"src/race.bash",
	"src/race.bat",
	"src/run.bash",
	"src/run.bat",
	"src/run.rc",

	"src/lib9",
	"src/libbio",
	"src/liblink",

	// TODO(adg): option to preserve pprof
	"pkg/tool/GOOS_GOARCH/pprof",
	"src/cmd/pprof",

	// TODO(adg): option to preserve gofmt
	"bin/gofmt",
	"src/cmd/gofmt",

	// TODO(adg): option to preserve docs
	"doc",
}
