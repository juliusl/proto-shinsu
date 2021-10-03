package registryclient

import (
	"net/url"

	"github.com/juliusl/shinsu/pkg/control"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func PullImages(node *control.Node, desc ocispec.Descriptor) (*control.StreamDescriptor, error) {
	return control.CreateStreamDescriptor(node,
		func() (*control.Address, *url.URL, error) {
			return nil, nil, nil
		},
		[]byte(desc.Digest), func(b []byte) ([]byte, error) {
			d := desc.Digest.Algorithm().FromBytes(b)
			return []byte(d), nil
		})
}
