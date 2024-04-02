package dbus

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/juju/errors"
	"github.com/nats-io/nats.go/micro"
)

var serviceMap map[string]micro.Service

func addDestination(dest string) error {
	if serviceMap == nil {
		serviceMap = make(map[string]micro.Service)
	}

	subjects := []string{dest, dest + "_" + nkey}

	for _, subject := range subjects {
		cfg := micro.Config{
			Name:       strings.ReplaceAll(strings.ReplaceAll(subject, ".", "_"), ":", "_"),
			Version:    "0.0.1", // todo is this required, does it have any meaning in this context?
			QueueGroup: nkey,
		}

		srv, err := micro.AddService(conn, cfg)
		if err != nil {
			return errors.Annotate(err, "failed to register service")
		}

		if err = registerObject(dest, "/", srv); err != nil {
			return err
		}

		serviceMap[dest] = srv
	}

	return nil
}

func removeDestination(dest string) error {
	srv, ok := serviceMap[dest]
	if !ok {
		// todo create a const error
		return errors.New("destination not found")
	}
	defer func() {
		delete(serviceMap, dest)
	}()

	return srv.Stop()
}

func registerObject(dest string, path dbus.ObjectPath, srv micro.Service) error {
	obj := dbusConn.Object(dest, path)
	node, err := introspect.Call(obj)
	if err != nil {
		return errors.Annotatef(err, "failed to introspect object: %v", obj)
	}

	subj := fmt.Sprintf("dbus.bus.%s", strings.ReplaceAll(dest, ".", "_"))
	if node.Name != "/" {
		subj = subj + strings.ReplaceAll(node.Name, "/", ".")
	}

	group := srv.AddGroup(subj)

	for _, iface := range node.Interfaces {
		iface := iface
		// node := node

		for _, prop := range iface.Properties {
			propName := fmt.Sprintf("%s.%s", iface.Name, prop.Name)

			annotations, err := json.Marshal(prop.Annotations)
			if err != nil {
				return errors.Annotate(err, "failed to marshal annotations to JSON")
			}

			err = group.AddEndpoint(
				prop.Name,
				propHandler(obj, propName),
				micro.WithEndpointMetadata(map[string]string{
					"Type":            "property",
					"Annotations":     string(annotations),
					"Property-Type":   prop.Type,
					"Property-Access": prop.Access,
				}),
			)

			if err != nil {
				return errors.Annotate(err, "failed to register property endpoint")
			}
		}

		for _, method := range iface.Methods {

			methodName := fmt.Sprintf("%s.%s", iface.Name, method.Name)

			args, err := json.Marshal(method.Args)
			if err != nil {
				return errors.Annotate(err, "failed to marshal []Args to JSON")
			}

			annotations, err := json.Marshal(method.Annotations)
			if err != nil {
				return errors.Annotate(err, "failed to marshal []Annotations to JSON")
			}

			err = group.AddEndpoint(
				method.Name,
				methodHandler(obj, methodName, &method),
				micro.WithEndpointMetadata(map[string]string{
					"Type":        "method",
					"Args":        string(args),
					"Annotations": string(annotations),
				}),
			)

			if err != nil {
				return errors.Annotate(err, "failed to register method endpoint")
			}
		}
	}

	for _, child := range node.Children {

		childPath := string(path) + "/" + child.Name
		if path == "/" {
			// remove leading "/"
			childPath = childPath[1:]
		}

		if err := registerObject(dest, dbus.ObjectPath(childPath), srv); err != nil {
			return err
		}
	}

	return nil
}

func propHandler(obj dbus.BusObject, name string) micro.HandlerFunc {
	return func(req micro.Request) {
		prop, err := obj.GetProperty(name)
		if err != nil {
			_ = req.Error("100", err.Error(), nil)
			return
		}
		headers := micro.WithHeaders(
			micro.Headers{
				"NKey":      []string{nkey},
				"Signature": []string{prop.Signature().String()},
			},
		)
		_ = req.Respond([]byte(prop.String()), headers)
	}
}

func methodHandler(obj dbus.BusObject, name string, method *introspect.Method) micro.HandlerFunc {
	var inArgs []introspect.Arg
	for _, arg := range method.Args {
		if arg.Direction == "in" {
			inArgs = append(inArgs, arg)
		}
	}

	return func(req micro.Request) {
		flagHeader := req.Headers().Get("Method-Flag")
		if flagHeader == "" {
			flagHeader = "0"
		}

		flag, err := strconv.Atoi(flagHeader)
		if err != nil {
			_ = req.Error("100", "invalid 'Method-Flag' header", nil)
			return
		}

		// todo improve validation of args

		invokeArgs := make([]interface{}, len(inArgs))

		if len(req.Data()) > 0 {
			if err = json.Unmarshal(req.Data(), &invokeArgs); err != nil {
				_ = req.Error("100", "failed to unmarshal req.Data", nil)
				return
			}
		}

		call := obj.Call(name, dbus.Flags(flag), invokeArgs...)
		if call.Err != nil {
			_ = req.Error("100", "failed to invoke method: "+err.Error(), nil)
			return
		}

		bytes, err := json.Marshal(call.Body)
		if err != nil {
			_ = req.Error("100", "failed to marshal result to JSON", nil)
			return
		}

		headers := micro.WithHeaders(
			micro.Headers{
				"NKey": []string{nkey},
			},
		)
		_ = req.Respond(bytes, headers)
	}
}
