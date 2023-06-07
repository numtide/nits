package state

import "github.com/nats-io/nats.go"

var NarInfoConfig = nats.KeyValueConfig{
	Bucket: "nar-info",
}

var NarInfoAccessConfig = nats.KeyValueConfig{
	Bucket: "nar-info-access",
}

var DeploymentConfig = nats.KeyValueConfig{
	Bucket:  "deployment",
	History: 32,
}

func InitKeyValueStores(js nats.JetStreamContext) (err error) {
	_, err = js.CreateKeyValue(&NarInfoConfig)
	if err != nil {
		return err
	}

	_, err = js.CreateKeyValue(&NarInfoAccessConfig)
	if err != nil {
		return err
	}

	_, err = js.CreateKeyValue(&DeploymentConfig)
	return err
}

func NarInfo(js nats.JetStreamContext) (nats.KeyValue, error) {
	return js.KeyValue(NarInfoConfig.Bucket)
}

func NarInfoAccess(js nats.JetStreamContext) (nats.KeyValue, error) {
	return js.KeyValue(NarInfoAccessConfig.Bucket)
}

func Deployment(js nats.JetStreamContext) (nats.KeyValue, error) {
	return js.KeyValue(DeploymentConfig.Bucket)
}
