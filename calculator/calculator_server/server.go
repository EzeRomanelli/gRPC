package main

import (
	"context"
	"fmt"
	"gRPC/calculator/calculatorpb"
	"io"
	"log"
	"math"
	"net"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {
	fmt.Printf("Sum function was invoked with: %v", req)
	firstInteger := req.GetSum().GetFirstInteger()
	secondInteger := req.GetSum().GetSecondInteger()
	result := firstInteger + secondInteger
	res := &calculatorpb.SumResponse{
		Result: result,
	}
	return res, nil
}

func (*server) PrimeNumberDecomposition(req *calculatorpb.PrimeNumberDecompositionRequest, stream calculatorpb.CalculatorService_PrimeNumberDecompositionServer) error {
	fmt.Printf("Received PrimeNumberDecomposition RPC: %v", req.GetPrimenumber())

	primeNumber := req.GetPrimenumber()
	divisor := int64(2)

	for primeNumber > 1 {
		if primeNumber%divisor == 0 {
			stream.Send(&calculatorpb.PrimeNumberDecompositionResponse{
				PrimeFactor: divisor,
			})
			primeNumber = primeNumber / divisor
		} else {
			divisor++
			fmt.Println("Divisor has increased to: ", divisor)
		}
	}
	return nil
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Printf("Compute Average function was invoked with a streaming request")
	sum := int32(0)
	amountOfNumbers := 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			result := float64(sum) / float64(amountOfNumbers)
			// We have finished reading the client stream
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{
				Result: result,
			})
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
		}
		sum += req.GetNumber()
		amountOfNumbers++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Printf("FindMaximum function was invoked with a streaming request")
	maxNumber := int32(0)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			log.Fatalf("Error while reading client stream: %v", err)
			return err
		}
		receivedNumber := req.GetNumber()
		if receivedNumber > maxNumber {
			maxNumber = receivedNumber
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{
				Result: maxNumber,
			})
			if sendErr != nil {
				log.Fatalf("Error while sending data to client: %v", sendErr)
				return sendErr
			}
		}

	}
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Received SquareRoot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number),
		)
	}
	return &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil // The nil is the error
}

func main() {
	fmt.Println("Calculator Server Started")

	lis, err := net.Listen("tcp", "0.0.0.0:9080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
