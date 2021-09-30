package main

import (
	"fmt"

	"github.com/juliusl/shinsu/pkg/control"
)

func main() {
	for a, u := range routes {
		fmt.Printf("%s %s\n", format_record(a), u.String())
	}
}

func format_record(a *control.Address) string {
	h, r, n, t, ref := a.Record()

	return fmt.Sprintf("%s %s %s %s %s", h, r, n, t, ref)
}
