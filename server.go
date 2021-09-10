package main

import "github.com/ProjectAthenaa/task-loader-service/loader"

func main() {
	load := loader.NewLoader()
	load.Start()
}
