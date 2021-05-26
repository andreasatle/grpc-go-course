package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/andreasatle/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Hello, I'm a client!")

	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}

	defer connection.Close()

	c := greetpb.NewGreetServiceClient(connection)
	fmt.Printf("Client created: %v\n", c)

	//doUnary(c, "Andreas", "Atle")
	//doServerStreaming(c, "Antonius", "Atle")
	//doClientStreaming(c, []string{"Andreas", "Mellissa", "Antonius", "Annelie"})
	//doBiDiStreaming(c, []string{"Andreas", "Mellissa", "Antonius", "Annelie"})
	doUnaryWithDeadline(c, "Andreas", 5*time.Second)
	doUnaryWithDeadline(c, "Andreas", time.Second)
}

func doUnary(c greetpb.GreetServiceClient, firstName, lastName string) {
	fmt.Println("Starting a Unary Greet RPC...")
	req := &greetpb.GreetRequest{
		Greeting: NewGreeting(firstName, lastName),
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error calling Greet RPC: %v\n", err)
	}

	log.Printf("Response from Greet RPC: %v", res.GetResult())
}

func doServerStreaming(c greetpb.GreetServiceClient, firstName, lastName string) {
	fmt.Println("Starting a Server Streaming GreetManyTimes RPC...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: NewGreeting(firstName, lastName),
	}
	stream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error calling GreetManyTimes RPC: %v\n", err)
	}

	for {
		res, err := stream.Recv()

		// Check if server has closed connection (i.e. returned nil)
		if err == io.EOF {
			// Server done sending
			break
		}
		if err != nil {
			log.Fatalf("Error reading stream:, %v\n", err)
		}

		log.Printf("Resonse from GreetManyTimes: %v", res.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient, firstNames []string) {
	fmt.Println("Starting a Client Streaming LongGreet RPC...")

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error calling rpc LongGreet: %v", err)
		return
	}

	for _, firstName := range firstNames {
		req := &greetpb.LongGreetRequest{
			Greeting: NewGreeting(firstName, ""),
		}
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error receiving from LongGreet")
		return
	}
	fmt.Println(res)
}

func doBiDiStreaming(c greetpb.GreetServiceClient, firstNames []string) {
	fmt.Println("Starting a Bi-Directional Streaming GreetAll RPC...")

	stream, err := c.GreetAll(context.Background())
	if err != nil {
		log.Fatalf("Error calling rpc GreetAll: %v\n", err)
		return
	}

	waitc := make(chan struct{})

	go func() {
		for _, firstName := range firstNames {
			req := &greetpb.GreetAllRequest{Greeting: NewGreeting(firstName, "")}
			stream.Send(req)
			time.Sleep(200 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err == io.EOF {
				log.Printf("Error receiving data from server:", err)
				break
			}
			fmt.Println("Received from server:", res)
		}
		close(waitc)
	}()

	<-waitc
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, firstName string, timeout time.Duration) {
	fmt.Println("Starting a Unary GreetWithDeadline RPC...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: NewGreeting(firstName, ""),
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Deadline exceeded!")
			} else {
				fmt.Println("Unexpected error:", statusErr)
			}
		} else {
			log.Fatalf("Error calling GreetWithDeadline RPC: %v\n", err)
		}
		return
	}

	log.Printf("Response from GreetWithDeadline RPC: %v", res.GetResult())

}
func NewGreeting(firstName, lastName string) *greetpb.Greeting {
	return &greetpb.Greeting{
		FirstName: firstName,
		LastName:  lastName,
	}
}
