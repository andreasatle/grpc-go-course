package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/andreasatle/grpc-go-course/blog/blogpb"
	"github.com/andreasatle/grpc-go-course/blog/server/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

var collection *mongo.Collection

// server implements the GreetServiceServer interface.
// This is confusing, since there are more than one of GreetManyTimes in the auto-generated file.
type server struct{}

// CreateBlog is an RPC for the Blog Service to create an entry in the database
func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	log.Println("Invoked RPC CreateBlog...")
	// Get blog from request
	blog := req.GetBlog()

	// Create a database item
	data := database.BlogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	// Insert database item in MongoDB
	mongoRes, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v\n", err))
	}

	// Check that we retrieve an id
	oid, ok := mongoRes.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Error converting oid"))
	}

	// Return a response containing the full blog item
	res := &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}

	return res, nil
}

// ReadBlog is an RPC for the Blog Service to read an entry from the database
func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	log.Println("Invoked RPC ReadBlog...")
	// Get the blog_id from the request
	blogID := req.GetBlogId()

	// Convert to a database id
	oid, err := primitive.ObjectIDFromHex(blogID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID: %v", err))
	}
	// prepare a database item
	data := &blogItem{}
	filter := bson.M{"_id": oid}
	err = collection.FindOne(context.Background(), filter).Decode(data)
	if err != nil {
		log.Printf("Error retrieving data from database: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Error retrieving data from database: %v", err))
	}

	res := &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}
	return res, nil
}

// UpdateBlog is an RPC for the Blog Service to update an entry in the database
func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	log.Println("Invoked RPC UpdateBlog...")
	// Get the blog_id from the request
	blog := req.GetBlog()

	// Convert to a database id
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID: %v", err))
	}
	// prepare a database item
	data := &blogItem{}
	filter := bson.M{"_id": oid}
	err = collection.FindOne(context.Background(), filter).Decode(data)
	if err != nil {
		log.Printf("Error retrieving data from database: %v", err)
		return nil, err
	}

	data.AuthorID = blog.GetAuthorId()
	data.Title = blog.GetTitle()
	data.Content = blog.GetContent()

	_, err = collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		log.Printf("Error updating data in database: %v", err)
		return nil, err
	}

	return &blogpb.UpdateBlogResponse{Blog: dataToBlogPb(data)}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	log.Println("Invoked RPC DeleteBlog...")
	oid, err := primitive.ObjectIDFromHex(req.GetBlogId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID: %v", err))
	}
	filter := bson.M{"_id": oid}
	_, err = collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting data in database: %v", err)
		return nil, err
	}
	return &blogpb.DeleteBlogResponse{}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	log.Println("Invoked RPC ListBlog...")
	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error I: %v", err))
	}
	defer cursor.Close(context.Background())
	data := &blogItem{}
	for cursor.Next(context.Background()) {
		err := cursor.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Error decoding data: %v", err))
		}
		stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlogPb(data)})
	}
	if err := cursor.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error II: %v", err))
	}
	return nil
}

func dataToBlogPb(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}
func main() {
	// Setup the logging, for if program crashes
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Setup MongoDB
	log.Println("Create a context...")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer func() {
		log.Println("Cancel Context...")
		cancel()
	}()

	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.Println("Shutting down MongoDB...")
		client.Disconnect(context.TODO())
	}()

	collection = client.Database("mydb").Collection("blog")
	// Start a tcp listener
	log.Println("Listen to tcp...")
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Printf("Failed to listen at tcp: %v\n", err)
	}
	defer func() {
		log.Println("Close the tcp-listener...")
		listener.Close()
	}()

	tls := false
	opts := []grpc.ServerOption{}
	if tls {
		certFile := "tsl/server.crt"
		keyFile := "tsl/server.key"

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
	defer func() {
		log.Println("Stop the gRPC server...")
		s.Stop()
	}()

	log.Println("Starting Blog service")
	// Register service
	blogpb.RegisterBlogServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		log.Println("Serve the gRPC server with tcp-listener...")
		// Serve service
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()

	// Wait for Contol-C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until signal is received
	<-ch
	log.Println("Shutting down server...")
}
