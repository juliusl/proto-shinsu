package control

import (
	"net/http"
	"net/url"
)

func CreateNode(address *Address, transport *NodeTransport) (*Node, error) {
	n := &Node{
		address:   address,
		transport: transport,
	}

	add, err := n.address.SetHost(transport.cache.name)
	if err != nil {
		return nil, err
	}

	n.address = add
	return n, nil
}

type Node struct {
	address   *Address
	transport *NodeTransport
}

var _ (http.CookieJar) = (*Node)(nil)

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (n *Node) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if n.address.host != u.Host {
		return
	}

	switch u.Scheme {
	case "http", "https":
	case "api":
	case "cache", "file":
	}
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (n *Node) Cookies(u *url.URL) []*http.Cookie {
	return nil
}

func (n *Node) GetClient() (*http.Client, error) {
	c := &http.Client{
		Transport: n.transport,
		Jar:       n,
	}

	return c, nil
}
