package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"github.com/andreasatle/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type server struct{}

func (s *server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Server Sum function invoked with req: %v\n", req)
	sum := int32(0)
	for _, num := range req.GetNums() {
		sum += num
	}
	return &calculatorpb.SumResponse{Sum: sum}, nil
}

func (s *server) PrimeNumber(req *calculatorpb.PrimeNumberRequest, stream calculatorpb.CalculatorService_PrimeNumberServer) error {
	fmt.Printf("Server PrimeNumber function invoked with req: %v\n", req)
	num := req.GetNum()
	for k := int32(2); num >= 2; {
		if num%k == 0 {
			res := &calculatorpb.PrimeNumberResponse{
				Prime: k,
			}
			stream.Send(res)
			time.Sleep(time.Second)
			num /= k
		} else {
			k++
		}
	}
	return nil
}

func (s *server) Average(stream calculatorpb.CalculatorService_AverageServer) error {
	fmt.Println("Server Average function invoked")
	sum, cnt := int32(0), int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&calculatorpb.AverageResponse{
				Average: float32(sum) / float32(cnt),
			})
		}
		if err != nil {
			log.Fatalf("Error retrieving data from client: %v", err)
			return err
		}
		sum += req.GetNum()
		cnt++
		log.Println("Received data from client:", req)
	}
}

func (s *server) FindMax(stream calculatorpb.CalculatorService_FindMaxServer) error {
	fmt.Println("Server FindMax function invoked")
	currMax := int32(-math.MaxInt32)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error receiving data from client: %v\n", err)
			return err
		}
		num := req.GetNum()
		fmt.Println("Received number:", num)
		if num > currMax {
			currMax = num
			fmt.Println("Sending new Max:", currMax)
			err := stream.Send(&calculatorpb.FindMaxResponse{Max: currMax})
			if err != nil {
				log.Fatalf("Error sending data to client: %v\n", err)
				return err
			}
		}
	}
}

func (s *server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Printf("Server SquareRoot function invoked with req: %v\n", req)
	num := req.GetNum()
	if num < 0.0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative argument: %v", num))
	}
	sqrt := math.Sqrt(num)
	return &calculatorpb.SquareRootResponse{Sqrt: sqrt}, nil
}

func main() {
	fmt.Println("Hello, I'm serving a calculator!")

	// Start a tcp listener
	listener, err := net.Listen("tcp", "0.0.0.0:50052")
	if err != nil {
		log.Fatalf("Failed to listen at tcp: %v\n", err)
	}

	// Create a new server
	s := grpc.NewServer()

	// Register service
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)

	// Serve service
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v\n", err)
	}

}
