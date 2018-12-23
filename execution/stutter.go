package execution

import (
	"github.com/hyperledger/burrow/storage"
)

// Critical block 481222 (no txs after 477561)
const StutterHeight uint64 = 480000
const StutterBy = 2

var stutterVersion int64

func init() {
	stutterVersion = VersionAtHeight(StutterHeight-1)
}

func stutterSave(tree *storage.RWTree) (hash []byte, version int64, err error) {
	saves := 1
	if tree.Version() == stutterVersion {
		saves += StutterBy
	}
	for i := 0; i < saves; i++ {
		hash, version, err = tree.Save()
	}
	return
}

func VersionAtHeight(height uint64) int64 {
	version := int64(height) + VersionOffset
	if height >= StutterHeight {
		return version + StutterBy
	}
	return version
}
