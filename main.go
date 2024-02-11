package main

import (
	"github.com/ghoshRitesh12/montre/lib"
)

func main() {
	montre := lib.Init()
	montre.StartWatching()
}
