package main

import (
	"log"
	"os"
)

func main() {
	file, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	ociterm := NewOciTerm()
	ociterm.Run()
	defer file.Close()
}
