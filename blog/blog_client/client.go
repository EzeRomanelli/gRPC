package main

import (
	"context"
	"fmt"
	"gRPC/blog/blogpb"
	"gRPC/greet/greetpb"
	"io"
	"log"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Blog Client")
	cc, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer cc.Close()

	c := blogpb.NewBlogServiceClient(cc)

	// Create Blog
	fmt.Println("Creating the blog")
	blog := &blogpb.Blog{
		AuthorId: "Eze",
		Title:    "My First Blog",
		Content:  "Content of my first Blog",
	}

	createBlogRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{
		Blog: blog,
	})

	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Printf("Blog has been created: %v \n", createBlogRes)
	blogId := createBlogRes.GetBlog().GetId()

	// Reading the Blog
	fmt.Println("Reading the Blog\n")
	_, err2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "asdasda"})
	if err2 != nil {
		fmt.Printf("Error Happened while reading: %v \n", err2)
	}

	readBlogReq := &blogpb.ReadBlogRequest{
		BlogId: blogId,
	}

	readBlogRes, readBlogErr := c.ReadBlog(context.Background(), readBlogReq)
	if readBlogErr != nil {
		fmt.Printf("Error Happened while reading: %v \n", readBlogErr)
	}

	fmt.Printf("Blog was read: %v \n", readBlogRes)

	// Update Blog
	fmt.Println("Updating the Blog\n")

	newBlog := &blogpb.Blog{
		Id:       blogId,
		AuthorId: "Changed Author",
		Title:    "Editing my first title Blog",
		Content:  "Edited Content on my first Blog ",
	}

	updateRes, updateErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newBlog})

	if updateErr != nil {
		fmt.Printf("Error while updating: %v \n", updateErr)
	}
	fmt.Printf("Blog was updated: %v \n", updateRes)

	// Delete the blog
	fmt.Println("Deleting the Blog\n")
	deleteRes, deleteErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogId})
	if deleteErr != nil {
		fmt.Printf("Error while deleting: %v \n", deleteErr)
	}
	fmt.Printf("Blog was deleted: %v \n", deleteRes)

	// List Blog
	fmt.Println("List all the Blogs\n")
	blogStream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})

	if err != nil {
		log.Fatalf("Error while calling ListBlog RPC: %v", err)
	}
	for {
		msg, err := blogStream.Recv()
		if err == io.EOF {
			// Reached the end of the stream
			break
		}
		if err != nil {
			log.Fatalf("Error while reading stream: %v", err)
		}
		fmt.Printf("Response from ListBlog: %v", msg.Blog)
	}
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a Unary RPC...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Eze",
			LastName:  "Romanelli",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}
