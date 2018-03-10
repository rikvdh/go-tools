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
func addTodo(client pb.TodoClient, todo *pb.TodoRequest) error {
	ret, err := client.Add(context.Background(), todo)
	if err != nil {
		return err
	}
	if !ret.Success {
		return fmt.Errorf("unknown error occurred")
	}
	return nil
}

func markDone(client pb.TodoClient, id int) {
	err := addTodo(client, &pb.TodoRequest{
		Created: uint64(time.Now().Unix()),
		List:    viper.GetString("list"),
		Id:      int32(id),
		Done:    true,
	})
	if err != nil {
		logrus.Errorf("Problem marking item as done: %v", err)
		return
	}
	logrus.Infof("Item %d marked as done", id)
}

func markUndone(client pb.TodoClient, id int) {
	err := addTodo(client, &pb.TodoRequest{
		Created: uint64(time.Now().Unix()),
		List:    viper.GetString("list"),
		Id:      int32(id),
		Done:    false,
	})
	if err != nil {
		logrus.Errorf("Problem marking item as undone: %v", err)
		return
	}
	logrus.Infof("Item %d marked as undone", id)
}

func addOrEditItem(client pb.TodoClient, todo []string, editID int) {
	err := addTodo(client, &pb.TodoRequest{
		Id:      int32(editID),
		Created: uint64(time.Now().Unix()),
		List:    viper.GetString("list"),
		Title:   strings.Join(todo, " "),
		Done:    false,
	})

	if err != nil {
		if editID > 0 {
			logrus.Errorf("Problem editing item with id %d: %v", editID, err)
		} else {
			logrus.Errorf("Problem adding item: %v", editID, err)
		}
		return
	}
	if editID > 0 {
		logrus.Infof("Successfully edited item with ID: %d", editID)
	} else {
		logrus.Infof("Successfully added item")
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
		markDone(client, *doneID)
	} else if *undoneID > 0 {
		markUndone(client, *undoneID)
	} else if flag.NArg() >= 1 {
		addOrEditItem(client, flag.Args(), *editID)
	} else {
		todoList(client, &pb.TodoFilter{
			List: viper.GetString("list"),
			All:  *all,
		})
	}
}
