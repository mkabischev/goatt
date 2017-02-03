package main

import (
	"fmt"
	"io/ioutil"
)

func dummy() {

}

func loadFile(filename string) ([]byte, error) {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("file [%s] read error [%s]\n", filename, err)
		return []byte(""), err
	}
	return contents, nil
}

func readFiles(args []string) [][]byte {
	payloads := [][]byte{}
	for i, path := range args {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			fmt.Printf("[%05d] file [%s] read error [%s]\n", i, path, err)
			continue
		}
		payloads = append(payloads, data)
	}
	return payloads
}
