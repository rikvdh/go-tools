package main

import (
	"io"
	"log"
	"fmt"
	"flag"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/rikvdh/go-tools/lib/todo"
)

const (
	address = "localhost:50051"
)

// addTodo calls the RPC method CreateTodo of TodoServer
func addTodo(client pb.TodoClient, todo *pb.TodoRequest) {
	resp, err := client.Add(context.Background(), todo)
	if err != nil {
		log.Fatalf("Could not create Todo: %v", err)
	}
	if resp.Success {
		log.Printf("A new Todo has been added with id: %d", resp.Id)
	}
}

// getTodos calls the RPC method GetTodos of TodoServer
func getTodos(client pb.TodoClient, filter *pb.TodoFilter) {
	// calling the streaming API
	stream, err := client.List(context.Background(), filter)
	if err != nil {
		log.Fatalf("Error on get todos: %v", err)
	}
	for {
		todo, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.GetTodos(_) = _, %v", client, err)
		}

		if todo.Done {
			fmt.Printf("\u2611 % 6d. %s\n", todo.Id, todo.Title)
		} else {
			fmt.Printf("\u2610 % 6d. %s\n", todo.Id, todo.Title)
		}
	}
}
func main() {
	doneId := flag.Int("done", 0, "Add todo-id to mark as done")
	flag.Parse()

	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoClient(conn)

	if *doneId > 0 {
		addTodo(client, &pb.TodoRequest{
			Id: int32(*doneId),
			Done: true,
		})
		return
	}

	if flag.NArg() >= 1 {
		addTodo(client, &pb.TodoRequest{
			Title: strings.Join(flag.Args(), " "),
			Done: false,
		})
		return
	}

	// Filter with an empty Keyword
	filter := &pb.TodoFilter{Text: ""}
	getTodos(client, filter)
}
