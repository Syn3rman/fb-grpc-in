package main

import "C"
import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/calyptia/plugin"
)

var messages = make(chan plugin.Message, 50)

func init() {
	plugin.RegisterInput("go-test-input-plugin", "Golang input plugin for testing", &grpcPlugin{})
}

type grpcPlugin struct {
	host string
	log  plugin.Logger
}

func (plug *grpcPlugin) Init(ctx context.Context, fbit *plugin.Fluentbit) error {
	plug.host = fbit.Conf.String("localhost")
	plug.log = fbit.Logger
	return nil
}

func (plug grpcPlugin) Collect(ctx context.Context, ch chan<- plugin.Message) error {
	plug.log.Info("[grpc] collect")

	go serve_tcp()

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			if err != nil && !errors.Is(err, context.Canceled) {
				plug.log.Error("[grpc] operation failed")

				return err
			}

		case message := <-messages:
			plug.log.Info("[grpc] message received")
			ch <- message

			return nil
		}
	}
}

func serve_tcp() {

	listen, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	// close listener
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {

	// incoming request
	buffer := make([]byte, 1024)
	_, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	// write data to response

	time := time.Now()
	entry := plugin.Message{Time: time, Record: buffer}

	messages <- entry

	responseStr := fmt.Sprintf("Your message is: %v. Received time: %v", string(buffer[:]), time)
	conn.Write([]byte(responseStr))

	// close conn
	conn.Close()
}

func main() {

}
