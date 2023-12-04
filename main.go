package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"golang.org/x/mod/modfile"
)

func main() {
	var filterList, path string
	var isReplaced bool
	flag.StringVar(&path, "path", "go.mod", "set go.mod path")
	flag.StringVar(&filterList, "filters", "", "filters to apply to the list of files")
	flag.BoolVar(&isReplaced, "replace", true, "warn about replace directives")
	flag.Parse()

	content, err := os.ReadFile(path)
	failOnError(err, fmt.Sprintf("read go.mod at %s", path))

	fs, err := modfile.Parse("", content, nil)
	failOnError(err, "parse go.mod")

	// sort direct first
	sort.Slice(fs.Require, func(i, j int) bool {
		if fs.Require[i].Indirect == fs.Require[j].Indirect {
			return fs.Require[i].Mod.Path < fs.Require[j].Mod.Path
		}
		return !fs.Require[i].Indirect
	})

	fmt.Println("\n# Require:")
	for _, f := range fs.Require {
		fmt.Printf("%s@%s (%s)\n", f.Mod.Path, f.Mod.Version, indirect(f.Indirect))
	}

	if len(filterList) > 0 {
		var filters []string
		filters = strings.Split(filterList, ",")

		fmt.Println("\n# Filters:")
		for _, f := range fs.Require {
			var found bool
			for _, filter := range filters {
				if strings.HasPrefix(f.Mod.Path, filter) {
					found = true
					break
				}
			}
			if found {
				fmt.Printf("%s@%s (%s)\n", f.Mod.Path, f.Mod.Version, indirect(f.Indirect))
			}
		}
	}

	if isReplaced {
		fmt.Println("\n# Replace:")
		for _, f := range fs.Replace {
			fmt.Printf("%s => %s %v\n", f.Old.Path, f.New.Path, f.Old.Version)
		}
	}
}

type indirect bool

func (i indirect) String() string {
	if i {
		return "indirect"
	}
	return "direct"
}

func failOnError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", msg, err)
		os.Exit(1)
	}
}
