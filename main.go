package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	vis "vis/src"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . files")
		os.Exit(1)
	}
	targetDir := os.Args[1]

	info, err := vis.ApplyDir(targetDir)
	if err != nil {
		panic(err)
	}

	file, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(path.Join(targetDir, "files.json"), file, 0644)
	if err != nil {
		panic(err)
	}
}
