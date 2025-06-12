package main

import (
	"fmt"
	"log"

	"github.com/open-cmi/gmiko"
)

func main() {
	d, err := gmiko.NewDevice("h3c", "comware", "host", 22, "username", "password")
	if err != nil {
		log.Fatalf("gmiko new device failed: %s\n", err.Error())
	}

	err = d.Connect(3)
	if err != nil {
		log.Fatalf("gmiko connect device failed: %s\n", err.Error())
	}
	defer d.Disconnect()

	v, err := d.RunCommand("display version")
	if err != nil {
		log.Fatalf("gmiko run command failed: %s\n", err.Error())
	}
	fmt.Printf("%s\n", string(v))
}
