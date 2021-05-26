package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/andreasatle/grpc-go-course/calculator/calculatorpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello, I'm a client!")

	connection, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}

	defer connection.Close()

	c := calculatorpb.NewCalculatorServiceClient(connection)
	fmt.Printf("Client created: %v\n", c)

	// Maybe it's bad to return results, if using go-routines etc.
	//sum := doUnary(c, []int32{1, 2, 3, 5, 8})
	//primes := doServerStreaming(c, int32(471200))
	//average := doClientStreaming(c, []int32{1, 2, 3, 4, 5, 6, 7, 8, 9})
	max := doBiDiStreaming(c, []int32{1, 3, 2, 5, 8, 7, 6, 9})
	//fmt.Println(sum, primes, average, max)
	fmt.Println(max)
}

func doUnary(c calculatorpb.CalculatorServiceClient, nums []int32) int32 {
	fmt.Println("Calling a Unary Sum RPC...")
	req := &calculatorpb.SumRequest{
		Nums: nums,
	}

	res, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatalf("Error calling Sum RPC: %v\n", err)
	}

	sum := res.GetSum()
	log.Printf("Response from Sum RPC: %v", sum)
	return sum
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient, num int32) []int32 {
	fmt.Println("Calling a Server Streaming PrimeNumber RPC...")
	primes := []int32{}
	req := &calculatorpb.PrimeNumberRequest{
		Num: num,
	}
	stream, err := c.PrimeNumber(context.Background(), req)
	if err != nil {
		log.Fatalf("Error calling PrimeNumber RPC: %v\n", err)
		return primes
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
		prime := res.GetPrime()
		log.Printf("Resonse from PrimeNumber: %v", prime)
		primes = append(primes, prime)
	}
	return primes

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient, nums []int32) float32 {
	fmt.Println("Calling a Client Streaming Average RPC...")

	stream, err := c.Average(context.Background())
	if err != nil {
		log.Fatalf("Error calling rpc Average: %v", err)
		return 0.0
	}
	for _, num := range nums {
		req := &calculatorpb.AverageRequest{
			Num: num,
		}
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error receiving from rpc Average")
		return 0.0
	}
	fmt.Println("Average:", res)
	return res.GetAverage()

}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient, nums []int32) int32 {
	fmt.Println("Starting a Bi-Directional Streaming FindMax RPC...")
	stream, err := c.FindMax(context.Background())
	if err != nil {
		log.Fatalf("Error calling rpc FindMax: %v\n", err)
		return 0
	}

	var max int32

	waitc := make(chan struct{})

	go func() {
		for _, num := range nums {
			req := &calculatorpb.FindMaxRequest{Num: num}
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
			if err != nil {
				log.Fatalf("Error receiving data from server: %v", err)
				break
			}
			fmt.Println("Received new Max from server:", res)
			max = res.GetMax()
		}
		close(waitc)
	}()

	<-waitc
	return max
}
