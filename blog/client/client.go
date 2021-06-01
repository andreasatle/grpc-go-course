package main

import (
	"context"
	"fmt"
	"log"

	"github.com/andreasatle/grpc-go-course/blog/blogpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	// Setup the logging, for if program crashes
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Hello, I'm a client!")

	tls := true
	opts := grpc.WithInsecure()
	fmt.Printf("%T\n", opts)
	if tls {
		certFile := "tsl/server.crt"
		creds, err := credentials.NewClientTLSFromFile(certFile, "")
		if err != nil {
			log.Fatalf("Error loading credentials: %v", err)
			return
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	connection, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Could not connect: %v\n", err)
	}

	defer connection.Close()

	c := blogpb.NewBlogServiceClient(connection)
	fmt.Printf("Client created: %v\n", c)
	tst(c, "Andreas", "First", "first")
	tst(c, "Anton", "Second", "second")
}

func tst(c blogpb.BlogServiceClient, author, title, content string) {
	blog := &blogpb.Blog{
		AuthorId: author,
		Title:    title,
		Content:  content,
	}

	// Create an entry in the database on the server side
	createRes, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})
	if err != nil {
		log.Fatalf("Error receiving data from server: %v\n", err)
		return
	}
	log.Printf("Blog has been created: %v\n", createRes.Blog)

	// Read the entry that just was created from the database on the server side
	readRes, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: createRes.Blog.GetId()})
	if err != nil {
		log.Fatalf("Error receiving data from server: %v", err)
		return
	}
	log.Printf("Blog has been read: %v\n", readRes.Blog)
	updateBlog := &blogpb.Blog{
		Id:       createRes.GetBlog().GetId(),
		AuthorId: createRes.GetBlog().GetAuthorId() + "(mod)",
		Title:    createRes.GetBlog().GetTitle() + "{mod}",
		Content:  createRes.GetBlog().GetContent() + "[mod]",
	}

	// Update the entry that just was created from the database on the server side
	updateRes, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: updateBlog})
	if err != nil {
		log.Fatalf("Error updating data on server: %v", err)
		return
	}
	log.Printf("Blog has been updated: %v\n", updateRes.Blog)

	_, err = c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: createRes.GetBlog().GetId()})
	if err != nil {
		log.Fatalf("Error deleting data on server: %v", err)
		return
	}
	log.Printf("Blog has been deleted\n")
}
