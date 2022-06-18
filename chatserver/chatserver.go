package chatserver

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// A messge Unit
type messageUnit struct {
	MessageTime       time.Time
	ClientName        string
	MessageBody       string
	MessageUniqueCode int
	ClientUniqueCode  int
}

// Message Hanlde Object with one Lock and
// No. of messages to be stroed!!

type messageHandle struct {
	MQue []messageUnit
	mu   sync.Mutex
}

var messageHandleObject = messageHandle{}

type ChatServer struct {
}

//Initial Point to Send an receive messages!
// Taking Services_ChatServiceServer interface{} with Send() and Recv()
func (is *ChatServer) ChatService(csi Services_ChatServiceServer) error {

	clientUniqueCode := rand.Intn(1e6)
	errch := make(chan error)

	// receive messages - init a go routine
	go receiveFromStream(csi, clientUniqueCode, errch)

	// send messages - init a go routine
	go sendToStream(csi, clientUniqueCode, errch)

	return <-errch

}

//receive messages
func receiveFromStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	/*   1. Running under infinite Loop to receive messages from Client.
	   mssg, err := csi_.Recv()
	   2. Locking message handlerObject.
	   3. adding each message in Queue (we can live wihout Queue as well!)
		 4. keep Queue to be stored in DB or somethig for later!!
	*/

	//implement an infintie loop
	for {
		mssg, err := csi_.Recv()
		if err != nil {
			log.Printf("Error in receiving message from client :: %v", err)
			errch_ <- err
		} else {

			messageHandleObject.mu.Lock() // Locked Object for a while

			messageHandleObject.MQue = append(messageHandleObject.MQue, messageUnit{
				MessageTime:       time.Now(), // Passing one or many
				ClientName:        mssg.Name,
				MessageBody:       mssg.Body,
				MessageUniqueCode: rand.Intn(1e8),
				ClientUniqueCode:  clientUniqueCode_,
			})

			log.Printf("Logging for Message-Handle %v", messageHandleObject.MQue[len(messageHandleObject.MQue)-1])

			messageHandleObject.mu.Unlock()

		}
	}
}

//send message
func sendToStream(csi_ Services_ChatServiceServer, clientUniqueCode_ int, errch_ chan error) {

	/*
		   1. So first thing is to run things in infiinite loop.
		   2. Lock the message Object.
		   3. If len(messaheHandleObject) is zero then Break.
			 Read required-Inforatmion from Message and
		   4. Now use FromServer{} created in Proto to send back to client.
			 // delete the message at index 0 after sending to receiver
			 5. messageHandleObject.MQue = messageHandleObject.MQue[1:]
	*/

	//implement a loop
	for {

		//loop through messages in MQue
		for {

			time.Sleep(500 * time.Millisecond)

			messageHandleObject.mu.Lock()

			if len(messageHandleObject.MQue) == 0 {
				messageHandleObject.mu.Unlock()
				break
			}

			SentTime := messageHandleObject.MQue[0].MessageTime
			senderUniqueCode := messageHandleObject.MQue[0].ClientUniqueCode
			senderName4Client := messageHandleObject.MQue[0].ClientName
			message4Client := messageHandleObject.MQue[0].MessageBody

			messageHandleObject.mu.Unlock()

			//send message to designated client (do not send to the same client)
			// Send() and Recv() are running in same loop, so anything with different Uinque-ID
			// Will be captured by Recv()
			if senderUniqueCode != clientUniqueCode_ {

				err := csi_.Send(&FromServer{Ts: timestamppb.New(SentTime), Name: senderName4Client, Body: message4Client})

				if err != nil {
					errch_ <- err
				}

				messageHandleObject.mu.Lock()

				if len(messageHandleObject.MQue) > 1 {
					messageHandleObject.MQue = messageHandleObject.MQue[1:] // delete the message at index 0 after sending to receiver
				} else {
					messageHandleObject.MQue = []messageUnit{}
				}
				messageHandleObject.mu.Unlock()

			}

		}
		time.Sleep(100 * time.Millisecond)
	}
}
