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

func ParseCookies(cookies []*http.Cookie) (*Address, *State, error) {
	// v2://sha256:a7bb12345f79fba6999132a5e3c796b37803adb14843d3de406b8218a725b0c6@registry-1.docker.io/library/ubuntu#blobs/uploads
	return nil, nil, nil
}

func MarshalCookies(add *Address, state *State) []*http.Cookie {
	return nil
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
	case "ref":
		a, err := n.address.SetNamespace(u.Path)
		if err != nil {
			return
		}
		a, err = a.SetReference(u.User.String())
		if err != nil {
			return
		}
		n.address = a
	case "api":
		add, state, err := ParseCookies(cookies)
		if err != nil {
			return
		}

		err = state.IsStable()
		if err != nil {
			return
		}

		r, err := add.String()
		if err != nil {
			return
		}

		c, err := n.GetClient()
		if err != nil {
			return
		}

		resp, err := c.Get(r.String())
		if err != nil {
			return
		}
		defer resp.Body.Close()

		resp, err = c.Post(u.String(), state.mediatype, resp.Body)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		sd, err := n.transport.api[u.Path].Update(resp.Request.Context(), state.mediatype, state.hash)
		if err != nil {
			return
		}

		n.transport.api[u.Path] = sd
	}
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (n *Node) Cookies(u *url.URL) []*http.Cookie {
	switch u.Scheme {
	case "api":
		sd, ok := n.transport.api[u.Path]
		if !ok {
			return nil
		}
		return MarshalCookies(sd.address, sd.state)
	}

	return nil
}

func (n *Node) GetClient() (*http.Client, error) {
	c := &http.Client{
		Transport: n.transport,
		Jar:       n,
	}

	return c, nil
}
