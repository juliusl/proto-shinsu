package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/juliusl/gopts/pkg/opts"
	"github.com/juliusl/shinsu/pkg/control"
)

func init() {
	addresses = make(map[options][]*control.Address)
	routes = make(map[*control.Address]*url.URL)

	if len(os.Args) <= 1 {
		format_help()
	}

	err := opts.Parse(Options, Usage)
	if err != nil {
		format_error(err)
	}
}

func handleOption(option, value string) options {
	switch option {
	case "n", "node":
		scope = node

		u, err := url.Parse(value)
		if err != nil {
			return options{}
		}

		host := u.User.Username()
		root := u.Scheme
		term := u.Host

		add := control.EmptyAddress()
		add.SetHost(host)
		add.SetRoot(root)
		add.SetTerm(term)

		addresses[scope] = append(addresses[scope], add)

		return node
	case "a", "api":
		scope = api
		add := control.EmptyAddress()
		add.SetRoot(value)

		addresses[scope] = append(addresses[scope], add)
		return api
	case "r", "reference":
		u, err := url.Parse(value)
		if err != nil {
			return options{}
		}

		for a := range routes {
			a.SetHost(u.Host)
			a.SetNamespace(u.Path)
			a.SetReference(u.User.Username())
		}

		return reference
	default:
		switch scope {
		case api:
			if value == "" && option != "" {
				u, err := url.Parse(option)
				if err != nil {
					return options{}
				}

				a, err := LastAddress(addresses[scope]).SetTerm("")
				if err != nil {
					return options{}
				}

				routes[a] = u
				return scope
			} else if value == "" {
				return options{}
			}

			u, err := url.Parse(value)
			if err != nil {
				return options{}
			}

			a, err := LastAddress(addresses[scope]).SetTerm(option)
			if err != nil {
				return options{}
			}

			routes[a] = u
			return scope
		}
		return scope
	}
}

var (
	node options = options{
		name:        "option: node\n",
		usage:       "usage: --node <address>\n", // This needs to output a new node address
		description: "Use this option to add a node to the control group\n",
	}
	api options = options{
		name:        "option: api\n",
		usage:       "usage: --api <root> <term> <location>\n",
		description: "Use this option to add an api route to the node transport\n",
	}
	reference options = options{
		name:        "option: reference\n",
		usage:       "usage: --reference ref://<reference>@<host>/<namespace>/",
		description: "Use this option to set the reference",
	}

	addresses map[options][]*control.Address
	routes    map[*control.Address]*url.URL
	scope     options
)

func LastAddress(a []*control.Address) *control.Address {
	l := len(a)
	return a[l-1]
}

func Options(option, value string) error {
	key := handleOption(option, value)

	if (key == options{}) {
		return Help()
	}

	return nil
}

func Help() error {
	indent := true
	return fmt.Errorf("Usage: backbone --node | -- api \nOptions:\n%s\n%s",
		node.help(indent),
		api.help(indent))
}

func Usage(option string) string {
	return handleOption(option, "").usage
}

type options struct {
	name        string
	usage       string
	description string
}

func (o options) help(indent bool) string {
	if indent {
		sep := `       `
		return fmt.Sprintf("%s%s%s%s%s%s", sep, o.name, sep, o.usage, sep, o.description)
	}
	return fmt.Sprintf("%s%s%s", o.name, o.usage, o.description)
}

func format_help() {
	format_error(Help())
}

func format_error(err error) {
	os.Stderr.WriteString(err.Error())
	os.Stderr.WriteString("\n")
	os.Exit(1)
}
