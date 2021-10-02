package main

import (
	"fmt"

	"github.com/juliusl/shinsu/pkg/control"
)

func main() {
	for a := range routes {
		print_address(a)
	}

	a, err := control.EmptyAddress().FromString("https://registry-1.docker.io/v2/library/ubuntu/blobs/uploads/sha256:a7bb12345f79fba6999132a5e3c796b37803adb14843d3de406b8218a725b0c6")
	if err != nil {
		panic(err.Error())
	}

	print_address(a)
}

func print_address(a *control.Address) {
	fmt.Printf("\n\n\n")
	u, err := a.APIRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.NodeRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.CacheRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.FileRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.RefRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.URI()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.HTTPSRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())

	u, err = a.HTTPRoot()
	if err != nil {
		panic(err.Error())
	}

	fmt.Println(u.String())
}

func format_record(a *control.Address) string {
	h, r, n, t, ref := a.Record()

	return fmt.Sprintf("%s %s %s %s %s", h, r, n, t, ref)
}
