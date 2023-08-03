package state

import "github.com/nats-io/nats.go"

var NarInfo nats.KeyValue

var NarInfoConfig = nats.KeyValueConfig{
	Bucket: "nar-info",
}

func InitKeyValueStores(js nats.JetStreamContext) (err error) {
	NarInfo, err = js.CreateKeyValue(&NarInfoConfig)
	return err
}
