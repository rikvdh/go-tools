package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/rikvdh/go-tools/lib/todo"
)

// addTodo calls the RPC method CreateTodo of TodoServer
func addTodo(client pb.TodoClient, todo *pb.TodoRequest) {
	_, err := client.Add(context.Background(), todo)
	if err != nil {
		log.Fatalf("Could not create todo-item: %v", err)
	}
}

// todoList retrieves and print a list of todo's
func todoList(client pb.TodoClient, filter *pb.TodoFilter) {
	stream, err := client.List(context.Background(), filter)
	if err != nil {
		logrus.Fatalf("Error: %v", err.Error())
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
			fmt.Printf("\u2611 % 6d. %s     (%s)\n", todo.Id, todo.Title, time.Unix(int64(todo.Created), 0).Format(time.Stamp))
		} else {
			fmt.Printf("\u2610 % 6d. %s     (%s)\n", todo.Id, todo.Title, time.Unix(int64(todo.Created), 0).Format(time.Stamp))
		}
	}
}

func main() {
	doneID := flag.Int("done", 0, "Todo ID to mark as done")
	undoneID := flag.Int("undone", 0, "Todo ID to mark as un-done (reset done flag)")
	editID := flag.Int("edit", 0, "ID of Todo item to edit")
	all := flag.Bool("all", false, "Show all (including done) items")
	flag.Parse()

	viper.SetEnvPrefix("td")
	viper.SetConfigName(".td")
	viper.AddConfigPath("$HOME")
	viper.SetDefault("uri", "localhost:50051")
	viper.SetDefault("list", "default")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if _, ok := err.(viper.ConfigParseError); ok {
		logrus.Fatalf("Error reading configuration: %v", err)
	}

	// Set up a connection to the gRPC server.
	conn, err := grpc.Dial(viper.GetString("uri"), grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewTodoClient(conn)

	if *doneID > 0 {
		addTodo(client, &pb.TodoRequest{
			List: viper.GetString("list"),
			Id:   int32(*doneID),
			Done: true,
		})
		return
	}
	if *undoneID > 0 {
		addTodo(client, &pb.TodoRequest{
			List: viper.GetString("list"),
			Id:   int32(*undoneID),
			Done: false,
		})
		return
	}

	if flag.NArg() >= 1 {
		addTodo(client, &pb.TodoRequest{
			Id:      int32(*editID),
			Created: uint64(time.Now().Unix()),
			List:    viper.GetString("list"),
			Title:   strings.Join(flag.Args(), " "),
			Done:    false,
		})
		return
	}

	todoList(client, &pb.TodoFilter{
		Text: "",
		List: viper.GetString("list"),
		All:  *all,
	})
}
