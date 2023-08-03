package state

import "github.com/nats-io/nats.go"

func Init(conn *nats.Conn) error {
	js, err := conn.JetStream()
	if err != nil {
		return err
	}

	if err = InitKeyValueStores(js); err != nil {
		return err
	}

	return InitObjectStores(js)
}
