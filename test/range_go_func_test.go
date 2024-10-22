package test

import (
	"fmt"
	"runtime"
	"testing"
)

type ServerOption func(configuration *ServerConfiguration) error

type ServerConfiguration struct {
	Name       string
	ListenAddr string
	ListenPort int
}

func WithName(name string) ServerOption {
	return func(configuration *ServerConfiguration) error {
		configuration.Name = name
		return nil
	}
}

func NewServer(serverOptions ...ServerOption) {
	var configuration ServerConfiguration
	for _, serverOption := range serverOptions {
		err := serverOption(&configuration)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("%+v\n", configuration)
}

func TestMemoryLeak(t *testing.T) {
	msg := make([]byte, 1024*1024*100)
	msgType := make([]byte, 5)
	printAlloc()
	copy(msgType, msg)
	runtime.GC()
	printAlloc()
}

func printAlloc() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%d KB\n", m.Alloc/1024)
}
