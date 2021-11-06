package main

import (
	"log"
	"os"
)

func main() {
	for i, arg := range os.Args {
		log.Println(i, arg)
	}
}