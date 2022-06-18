package chatserver

import (
	"context"
	"log"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type clienthandle struct {
	stream     Services_ChatServiceClient
	clientName string
}

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	RegisterServicesServer(server, &ChatServer{})

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

	client := NewServicesClient(conn)
	// Now we got Stream as Object where you could do Send() and Recv() things!!
	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}
	clientMessage := "This is Message from Client"
	ch := clienthandle{stream: stream}
	clientMessageBox := FromClient{
		Ts:   timestamppb.Now(),
		Name: ch.clientName,
		Body: clientMessage,
	}

	err = ch.stream.Send(&clientMessageBox)

	if err != nil {
		log.Printf("Error while sending message to server :: %v", err)

	}

}

func TestRecvFromStream(t *testing.T) {

	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := NewServicesClient(conn)
	// Now we got Stream as Object where you could do Send() and Recv() things!!
	stream, err := client.ChatService(context.Background())
	if err != nil {
		log.Fatalf("Failed to call ChatService :: %v", err)
	}
	ch := clienthandle{stream: stream}
	data, err := ch.stream.Recv()

	if err != nil {
		log.Printf("Error while sending message to server :: %v", err)
	}
	log.Println(data)

}
