# gRPC using protocol buffers

We will write three different subprojects that implements the different aspects of gRPC:
* Unary API
* Server streaming API
* Client streaming API
* Bi-Directional streaming API

The APIs are defined in a protocol buffer version 3 (```syntax = "proto3";```)

# Sub projects
* Greet
* Calculator
* Blog


## Sub project Greet
This project is completed in the lectures.

## Sub project Calculator
This sub project is implemented as homework.

## Sub project Blog
This is a more involved project using a CRUD-framework with MongoDB.

# go-code generation from the protocol buffers
We use a bash script ```configure.sh```
```
[ $# -eq 1 ]
    && protoc $1/$1pb/$1.proto --go_out=plugins=grpc:.
    || echo "Usage: $0 <name>"
```
which is called with 
```
./configure.sh [greet|calculator|blog]
```
in order to generate the go-code from the greet protocol buffer file.
we can replace ```greet``` with ```calculator``` to generate the calculator go-files. Once the ```blog``` project is up and running, this will be an option too.

# Error handling in gRPC
We implement an ```rpc SquareRoot``` that checks that the argument in positive.
```go
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
```
We create an error with ```status.Errorf```, with error code ```codes.InvalidArgument```. For a list of valid arguments, click [grpc.io](https://grpc.io/docs/guides/error/) and [avi.im](http://avi.im/grpc-errors/).

# gRPC Deadlines
The client can set a deadline by modifying the context in the RPC.
Replace the RPC call:
```go
res, err := c.GreetWithDeadline(context.Background(), req)
```
with
```go
ctx, cancel := context.WithTimeout(context.Background(), timeout)
defer cancel()
res, err := c.GreetWithDeadline(ctx, req)
```
where the timeout is a ```time.Duration```.
After we can check for the gRPC-error ```codes.DeadlineExceeded```.<br>
The full code on the client side:
```go
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
```
The code on the server side:
```go
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
```
The server simulates that the request takes 3 seconds, and checks if the client has timed out
once every second.

# SSL Security
The instructor has given us scripts to generate the certificate files. The files on github doesn't match the file in the videos, so I modified the github files to match the video. Moreover, there seems to be an issue with go-1.15 and forward regarding the certification from the instructor.
I have to add a GODEBUG-flag in order for it to work. I will have to come back to this at a later time.

On the ser ver side, we modify the code:
```go
	tls := true
	opts := []grpc.ServerOption{}
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatal("Error loading certificates: %v", err)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	// Create a new server
	log.Println("Start the gRPC server...")
	s := grpc.NewServer(opts...)
```
If ```tls``` is set to ```true```, then we use certification, otherwise not.

On the client side, we modify:
```go
	tls := true
	opts := grpc.WithInsecure()
	if tls {
		certFile := "ssl/ca.crt"
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			log.Fatalf("Error loading credentials: %v", err)
			return
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	connection, err := grpc.Dial("localhost:50051", opts)
```
The ```tls``` option has to match the server. When the certification is used, we need to start the client with:
```
GODEBUG=x509ignoreCN=0 go run blog/client/client.go
```
Very annoying!

# Reflection and Evans CLI
Add one line in the server main function:
```go
	// Register service
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})
	reflection.Register(s)
```

Now we can use Evans CLI and reflection to obtain what API a server provides. Observe that
the server has to be registered as indicated above.

```
evans -p 50051 -r
```
listens to a server at port :50051.

### Example 1 - List services
```
calculator.CalculatorService@127.0.0.1:50051> show service
+-------------------+-------------+--------------------+---------------------+
|      SERVICE      |     RPC     |    REQUEST TYPE    |    RESPONSE TYPE    |
+-------------------+-------------+--------------------+---------------------+
| CalculatorService | Sum         | SumRequest         | SumResponse         |
| CalculatorService | PrimeNumber | PrimeNumberRequest | PrimeNumberResponse |
| CalculatorService | Average     | AverageRequest     | AverageResponse     |
| CalculatorService | FindMax     | FindMaxRequest     | FindMaxResponse     |
| CalculatorService | SquareRoot  | SquareRootRequest  | SquareRootResponse  |
+-------------------+-------------+--------------------+---------------------+
```

### Example 2 - List messages
```
calculator.CalculatorService@127.0.0.1:50051> show message
+---------------------+
|       MESSAGE       |
+---------------------+
| AverageRequest      |
| AverageResponse     |
| FindMaxRequest      |
| FindMaxResponse     |
| PrimeNumberRequest  |
| PrimeNumberResponse |
| SquareRootRequest   |
| SquareRootResponse  |
| SumRequest          |
| SumResponse         |
+---------------------+
```

### Example 3 - call RPC
```
calculator.CalculatorService@127.0.0.1:50051> call SquareRoot
num (TYPE_DOUBLE) => 42
{
  "sqrt": 6.48074069840786
}

calculator.CalculatorService@127.0.0.1:50051> call SquareRoot
num (TYPE_DOUBLE) => -42
command call: rpc error: code = InvalidArgument desc = Received a negative argument: -42
```