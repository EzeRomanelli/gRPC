package main

import (
	"context"
	"fmt"
	"gRPC/calculator/calculatorpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Hello, im a Client")
	cc, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)
	//fmt.Printf("Created client: %f", c)
	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBiDiStreaming(c)

	doErrorUnary(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Unary RPC...")
	req := &calculatorpb.SumRequest{
		Sum: &calculatorpb.Sum{
			FirstInteger:  10,
			SecondInteger: 3,
		},
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling SUM RPC: %v", err)
	}
	log.Printf("Response from SUM: %v", res.Result)
}

func doServerStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a PrimeDecomposition Server Streaming RPC...")
	req := &calculatorpb.PrimeNumberDecompositionRequest{
		Primenumber: 12,
	}

	stream, err := c.PrimeNumberDecomposition(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling PrimeDecomposition RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(res.GetPrimeFactor())
	}
}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a Client Streaming Computed Average RPC...")

	requests := []*calculatorpb.ComputeAverageRequest{
		&calculatorpb.ComputeAverageRequest{
			Number: 1,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 2,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 3,
		},
		&calculatorpb.ComputeAverageRequest{
			Number: 4,
		},
	}

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling ComputeAverage: %v", err)
	}

	// We iterate over the slice and send each message individualy
	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(10000 * time.Millisecond)
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while Receiving response from ComputeAverage: %v", err)
	}
	fmt.Printf("ComputeAverage Response: %v\n", response)
}

func doBiDiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a BiDi Streaming RPC...")

	requests := []*calculatorpb.FindMaximumRequest{
		&calculatorpb.FindMaximumRequest{
			Number: 1,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 5,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 15,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 10,
		},
		&calculatorpb.FindMaximumRequest{
			Number: 4,
		},
	}

	// We Create a stream by invoking the client
	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	waitc := make(chan struct{})
	// We send a bunch of messages to the client (go routine)
	go func() {
		// function to send a bunch of mesages
		for _, req := range requests {
			fmt.Printf("Sending Message: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	// We receive a bunch of messages from the client (go routine)
	go func() {
		// function to receive a bunch of mesages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}
			fmt.Println("Received: ", res.GetResult())
		}
		close(waitc)
	}()

	// block until everything is done
	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a SquareRoot Unary RPC...")
	// Correct call
	doErrorCall(c, 10)

	// Error call
	doErrorCall(c, -2)
}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) {
	res, err := c.SquareRoot(context.Background(), &calculatorpb.SquareRootRequest{Number: n})

	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// Actual error from gRPC (user error)
			fmt.Printf("Error message from server: %v\n", respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number!")
				return
			}
		} else {
			log.Fatalf("Big Error calling SquareRoot: %v", err)
			return
		}
	}

	fmt.Printf("Result of square root of %v : %v\n", n, res.GetNumberRoot())
}
