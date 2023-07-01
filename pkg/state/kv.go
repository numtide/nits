package state

import "github.com/nats-io/nats.go"

var NarInfo nats.KeyValue

var NarInfoConfig = nats.KeyValueConfig{
	Bucket: "nar-info",
}

var Deployment nats.KeyValue

var DeploymentConfig = nats.KeyValueConfig{
	Bucket: "deployment",
	// TODO max history size is hardcoded to 64 in the go client but there isn't a limit enforced by the server
	History: 64,
}

var DeploymentResultConfig = nats.KeyValueConfig{
	Bucket:  "deployment-result",
	History: 64,
}

var DeploymentResult nats.KeyValue

func InitKeyValueStores(js nats.JetStreamContext) (err error) {
	NarInfo, err = js.CreateKeyValue(&NarInfoConfig)
	if err != nil {
		return err
	}

	Deployment, err = js.CreateKeyValue(&DeploymentConfig)
	if err != nil {
		return err
	}

	DeploymentResult, err = js.CreateKeyValue(&DeploymentResultConfig)
	return err
}
