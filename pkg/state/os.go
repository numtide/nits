package state

import "github.com/nats-io/nats.go"

var NarConfig = nats.ObjectStoreConfig{
	Bucket: "nar",
}

func InitObjectStores(js nats.JetStreamContext) (err error) {
	_, err = js.CreateObjectStore(&NarConfig)
	return err
}

func Nar(js nats.JetStreamContext) (nats.ObjectStore, error) {
	return js.ObjectStore(NarConfig.Bucket)
}
