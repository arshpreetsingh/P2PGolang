package main

import (
	"context"
	"grpcChatServer/chatserver"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// type clienthandle struct {
// 	stream     Services_ChatServiceClient
// 	clientName string
// }

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	chatserver.RegisterServicesServer(server, &chatserver.ChatServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestSendToStream(t *testing.T) {

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := chatserver.NewServicesClient(conn)
	// Now we got Stream as Object where you could do Send() and Recv() things!!
	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}
	//	clientMessage := "This is Message from Client"
	ch := clienthandle{stream: stream}

	ch.clientConfig()
	ch.sendMessage()
	ch.receiveMessage()
}
