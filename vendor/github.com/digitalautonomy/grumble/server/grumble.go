// Copyright (c) 2010 The Grumble Authors
// The use of this source code is goverened by a BSD-style
// license that can be found in the LICENSE-file.

package server

import (
	"github.com/digitalautonomy/grumble/pkg/blobstore"
)

var servers map[int64]*Server
var blobStore blobstore.BlobStore

func SetServers(s map[int64]*Server) {
	servers = s
}

func SetBlobStore(b blobstore.BlobStore) {
	blobStore = b
}
