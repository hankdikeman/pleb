/*
 * Client wrappers for metadata access.
 */

package netclient

import (
	// senator "github.com/pleb/prod/senator/pb"

	"bazil.org/fuse"
)

func (netclient *NetClient) GetAttr(handle fuse.HandleID) (*fuse.Attr, error) {
	// netclient.client.GetAttr(
	return nil, nil
}

func (netclient *NetClient) SetAttr(handle fuse.HandleID, attr *fuse.Attr) error {
	// netclient.client.SetAttr(
	return nil
}
