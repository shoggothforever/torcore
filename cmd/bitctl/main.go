/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
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
	//cmd.Execute()
	mp := make(map[string]string)
	mp["1"] = "2"
	mp["2"] = "2"
	mp["3"] = "2"
	mp["4"] = "2"
	mp["5"] = "2"
	fmt.Println("starting iter")
	for k, v := range OrderMaps(mp) {
		fmt.Println(k, v)
	}

}
