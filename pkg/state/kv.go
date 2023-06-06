package state

import "github.com/nats-io/nats.go"

var NarInfoConfig = nats.KeyValueConfig{
	Bucket: "nar-info",
}

var NarInfoAccessConfig = nats.KeyValueConfig{
	Bucket: "nar-info-access",
}

func InitKeyValueStores(js nats.JetStreamContext) (err error) {
	_, err = js.CreateKeyValue(&NarInfoConfig)
	if err != nil {
		return err
	}

	_, err = js.CreateKeyValue(&NarInfoAccessConfig)
	return err
}

func NarInfo(js nats.JetStreamContext) (nats.KeyValue, error) {
	return js.KeyValue(NarInfoConfig.Bucket)
}

func NarInfoAccess(js nats.JetStreamContext) (nats.KeyValue, error) {
	return js.KeyValue(NarInfoAccessConfig.Bucket)
}
