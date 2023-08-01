package main

import "fmt"
import "log"

var ARCH, OS, SEMVER string

func main() {
	log.Println("hello world!")
	log.Println("i was created by repo-template-go")
	log.Println(Version())
}

func Version() string {
	return fmt.Sprintf("%s-%s-%s", SEMVER, OS, ARCH)
}
