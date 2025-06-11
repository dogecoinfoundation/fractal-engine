package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	ips, err := net.LookupIP("dogenet-alpha")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		os.Exit(1)
	}
	for _, ip := range ips {
		fmt.Println(ip.String())
	}
}
