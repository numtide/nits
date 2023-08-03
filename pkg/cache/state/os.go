package state

import "github.com/nats-io/nats.go"

var Nar nats.ObjectStore

var NarConfig = nats.ObjectStoreConfig{
	Bucket: "nar",
}

func InitObjectStores(js nats.JetStreamContext) (err error) {
	Nar, err = js.CreateObjectStore(&NarConfig)
	return err
}
