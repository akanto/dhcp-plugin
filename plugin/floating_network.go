package dhcp

import (
	"encoding/json"

	"github.com/docker/libnetwork/datastore"

	log "github.com/Sirupsen/logrus"
)

type network struct {
	Id      string
	DbIndex uint64
	DbExists  bool
	Gateway string
}

// Key method lets an object to provide the Key to be used in KV Store
func (n *network) Key() []string {
	log.Debugf("Key invoked on network: %+v", n)
	return []string{"floating", "network", n.Id}
}

// KeyPrefix method lets an object to return immediate parent key that can be used for tree walk
func (n *network) KeyPrefix() []string {
	log.Debugf("KeyPrefix invoked on network: %+v", n)
	return []string{"floating", "network"}
}

// Value method lets an object to marshal its content to be stored in the KV store
func (n *network) Value() []byte {
	log.Debugf("Value invoked on network: %+v", *n)
	b, err := json.Marshal(*n)
	if err != nil {
		return []byte{}
	}
	log.Debugf("Marshalled, size: %d", len(b))
	return b
}

// SetValue is used by the datastore to set the object's value when loaded from the data store.
func (n *network) SetValue(value []byte) error {
	log.Debugf("SetValue invoked on network: %+v, value len: %d", n, len(value))
	storedNet := &network{}

	err := json.Unmarshal(value, storedNet)
	log.Infof("Value retrieved: %+v, err: %s", storedNet, err)
	return json.Unmarshal(value, n)
}

// Index method returns the latest DB Index as seen by the object
func (n *network) Index() uint64 {
	log.Debugf("Index invoked on network: %+v", n)
	return n.DbIndex
}

// SetIndex method allows the datastore to store the latest DB Index into the object

func (n *network) SetIndex(index uint64) {
	log.Debugf("SetIndex invoked on network: %+v", n)
	n.DbIndex = index
	n.DbExists = true
}

// True if the object exists in the datastore, false if it hasn't been stored yet.
// When SetIndex() is called, the object has been stored.
func (n *network) Exists() bool {
	log.Debugf("Exists invoked on network: %+v", n)
	return n.DbExists
}

// DataScope indicates the storage scope of the KV object
func (n *network) DataScope() string {
	log.Debugf("DataScope invoked on network: %+v", n)
	return datastore.GlobalScope
}

// Skip provides a way for a KV Object to avoid persisting it in the KV Store
func (n *network) Skip() bool {
	log.Debugf("Skip invoked on network: %+v", n)
	return false
}









