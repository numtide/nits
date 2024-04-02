package main

import (
	"github.com/charmbracelet/log"
	"github.com/godbus/dbus/v5"
)

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		log.Fatal(err)
	}

	obj := conn.Object("org.freedesktop.systemd1", "/org/freedesktop/systemd1")
	call := obj.Call("org.freedesktop.systemd1.Manager.ListUnits", 0)

	log.Info(call)

	//if err = bus.Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
	//	log.Fatal(err)
	//}
	//
	//for _, name := range names {
	//	log.Info(name)
	//	node, err := introspect.Call(conn.Object(name, "/"))
	//	if err != nil {
	//		log.Error(err)
	//	}
	//	log.Info(node)
	//}
}
