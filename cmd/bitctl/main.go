/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"github.com/shoggothforever/torcore/cmd/bitctl/cmd"
	"iter"
	"slices"
)

func OrderMaps(mp map[string]string) iter.Seq2[string, string] {
	keys := make([]string, len(mp))
	i := 0
	for k, _ := range mp {
		keys[i] = k
		i++
	}
	slices.Sort(keys)
	return func(yield func(string, string) bool) {
		for _, k := range keys {
			if !yield(k, mp[k]) {
				return
			}
		}
	}

}

func main() {
	cmd.Execute()
}
