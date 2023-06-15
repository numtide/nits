package main

import (
	"bufio"
	"bytes"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

type outLogger struct {
	buf     *bytes.Buffer
	scanner *bufio.Scanner
	logFn   func(msg string, ctx ...interface{})
}

func (o outLogger) Write(b []byte) (n int, err error) {
	n, err = o.buf.Write(b)
	for o.scanner.Scan() {
		msg := o.scanner.Text()
		clean := strings.Map(func(r rune) rune {
			if unicode.IsGraphic(r) {
				return r
			}
			return -1
		}, msg)
		o.logFn(clean)
	}
	return
}

func newOutLogger(logFn func(msg string, ctx ...interface{})) outLogger {
	buf := bytes.NewBuffer(make([]byte, 1024))
	scanner := bufio.NewScanner(buf)
	return outLogger{
		buf:     buf,
		scanner: scanner,
		logFn:   logFn,
	}
}

const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansi)

type foo struct {
	buf   []byte
	logFn func(msg string, ctx ...interface{})
}

func (o foo) Write(b []byte) (n int, err error) {
	buf := o.buf

	buf = append(buf, b...)
	scanner := bufio.NewScanner(bytes.NewBuffer(buf))

	count := 0

	for scanner.Scan() {
		msg := scanner.Text()
		count += len(msg)
		o.logFn(msg)
	}

	// resize based on number of bytes read
	o.buf = buf[count:]

	return len(b), nil
}

func newFoo(logFn func(msg string, ctx ...interface{})) foo {
	return foo{
		logFn: logFn,
	}
}

func main() {
	cmd := exec.Command("nix", "copy", "--to", " http://localhost:3000/\\?compression\\=zstd", "-v", "--log-format", "raw", "nixpkgs#hello")

	cmd.Stdout = newFoo(func(msg string, ctx ...interface{}) {
		println(msg)
	})
	cmd.Stderr = newFoo(func(msg string, ctx ...interface{}) {
		println("err: " + msg)
	})

	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
