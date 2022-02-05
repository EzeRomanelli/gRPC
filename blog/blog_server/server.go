package main

import (
	"context"
	"fmt"
	"gRPC/blog/blogpb"
	"log"
	"net"
	"os"
	"os/signal"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"

	"google.golang.org/grpc"
)

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

type blogItem struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Author  string             `bson:"author_id"`
	Content string             `bson:"content"`
	Title   string             `bson:"title"`
}

var collection *mongo.Collection
var ctx = context.TODO()

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {

	fmt.Println("Create log request")
	blog := req.GetBlog()

	data := blogItem{
		Author:  blog.GetAuthorId(),
		Title:   blog.GetTitle(),
		Content: blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error: %v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot convert to OID"),
		)
	}

	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil

}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read Blog request")
	blogID := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot Parse ID"),
		)
	}
	// create an empty struct
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}
	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.Author,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update Blog request")
	blog := req.GetBlog()
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot Parse ID"),
		)
	}
	data := &blogItem{}
	filter := bson.D{{"_id", oid}}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID: %v", err),
		)
	}

	data.Author = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()

	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if updateErr != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot update Object MongoDB: %v \n", updateErr),
		)
	}

	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogPb(data),
	}, nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.Author,
		Content:  data.Content,
		Title:    data.Title,
	}
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete Blog request")
	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot Parse ID"),
		)
	}
	filter := bson.D{{"_id", oid}}

	res, deleteErr := collection.DeleteOne(context.Background(), filter)
	if deleteErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete Object MongoDB: %v \n", deleteErr),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find Object to delete in MongoDB: %v \n", deleteErr),
		)
	}

	return &blogpb.DeleteBlogResponse{
		BlogId: req.GetBlogId(),
	}, nil

}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("List Blog request")

	filter := bson.D{}
	// Filter = nil to get All blogs
	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		return status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find any Blogs \n"),
		)
	}

	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Internal Error while decoding data: %v \n", err),
			)
		}
		stream.Send(&blogpb.ListBlogResponse{
			Blog: dataToBlogPb(data),
		})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v \n", err),
		)
	}

	return nil
}

func main() {
	// If we crash the Go code, we get the file and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog Service Started")

	fmt.Println("Connecting to MongoDB")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:9080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for Ctrl + C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	fmt.Println("End of program")

}
