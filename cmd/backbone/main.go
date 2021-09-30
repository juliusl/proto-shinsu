package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/juliusl/gopts/pkg/opts"
)

var (
	a, b string
)

func init() {
	err := opts.Parse(Options, Usage)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
}

func main() {

	fmt.Println(a, b)
}

func Options(option, value string) error {
	switch option {
	case "n", "node":
		_, err := url.Parse(value)
		if err != nil {
			return err
		}

	case "a", "api":
		if value == "test2" {
			return errors.New("test2 is not an allowed value")
		}
		b = value
	}

	return nil
}

func Usage(option string) string {
	return handleOption(option, "").usage
}

func handleOption(option, value string) options {
	switch option {
	case "n", "node":
		return node
	case "b", "ball":
		return api
	default:
		return options{}
	}
}

type options struct {
	name  string
	usage string
}

var (
	node options = options{name: "node", usage: "usage: --node <host>@<root>://<term>"}
	api  options = options{name: "api", usage: "usage: --api <root>`<term>` <location>"}
)
