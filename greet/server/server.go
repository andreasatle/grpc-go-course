package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/andreasatle/grpc-go-course/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// server implements the GreetServiceServer interface.
// This is confusing, since there are more than one of GreetManyTimes in the auto-generated file.
type server struct{}

func (s *server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	fmt.Printf("Server Greet function invoked with req: %v\n", req)

	res := &greetpb.GreetResponse{
		Result: "Hello " + req.GetGreeting().GetFirstName(),
	}
	return res, nil
}

func (s *server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Printf("Server GreetManyTimes function invoked with req: %v\n", req)

	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " " + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		err := stream.Send(res)
		if err != nil {
			log.Fatalf("Error sending data to client: %v\n", err)
			return err
		}

		time.Sleep(time.Second)
	}
	return nil
}

func (s *server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("Server LongGreet function invoked")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// Once client done, then return the result.
			return stream.SendAndClose(&greetpb.LongGreetResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error receiving stream from client: %v\n", err)
			return err
		}

		// retrieve next first name.
		firstName := req.GetGreeting().GetFirstName()
		log.Println("Received data from client:", req)
		result += "Hello " + firstName + "! "
	}

}

func (s *server) GreetAll(stream greetpb.GreetService_GreetAllServer) error {
	fmt.Println("Server GreetAll function invoked")

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error receiving data from client: %v\n", err)
			return err
		}
		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "!"
		err = stream.Send(&greetpb.GreetAllResponse{
			Result: result,
		})
		if err != nil {
			log.Fatalf("Error sending data to client: %v\n", err)
			return err
		}
	}
}

func (s *server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	fmt.Printf("Server GreetWithDeadline function invoked with req: %v\n", req)

	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			fmt.Println("The client canceled the request")
		}
		time.Sleep(time.Second)
	}
	res := &greetpb.GreetWithDeadlineResponse{
		Result: "Hello " + req.GetGreeting().GetFirstName(),
	}
	return res, nil
}

func main() {
	fmt.Println("Hello, I'm serving a greeting!")

	// Start a tcp listener
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen at tcp: %v\n", err)
	}

	// Create a new server
	s := grpc.NewServer()

	// Register service
	greetpb.RegisterGreetServiceServer(s, &server{})
	reflection.Register(s)

	// Serve service
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}
}
