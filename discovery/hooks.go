package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Hookset struct {
	PreBuild  []string
	PostBuild []string
}

func findHooks(source string) ([]string, bool, error) {
	dir, err := os.ReadDir(source)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, false, nil
		}
		return nil, false, err
	}

	var hooks []string

	for _, file := range dir {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".lua" {
			hooks = append(hooks, file.Name())
		}
	}

	sort.Strings(hooks) // hooks MUST be executed in alphabetical order, mainly for ordering with numbers
	// eg 1-x.lua, 2-y.lua, 3-z.lua.
	return hooks, true, nil
}

func DiscoverHooks(source string) (*Hookset, error) {
	preDir := filepath.Join(source, "pre")
	postDir := filepath.Join(source, "post")

	pre, preExists, err := findHooks(preDir)
	if err != nil {
		return nil, err
	}
	post, postExists, err := findHooks(postDir)
	if err != nil {
		return nil, err
	}

	if !(preExists || postExists) {
		return nil, fmt.Errorf("no hooks found, neither %q nor %q exist", preDir, postDir)
	}

	return &Hookset{pre, post}, nil
}
