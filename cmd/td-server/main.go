package main

import (
	"log"
	"net"
	"fmt"
	"encoding/binary"
	"encoding/json"
	//"strings"

	"golang.org/x/net/context"
	"github.com/boltdb/bolt"
	"google.golang.org/grpc"

	pb "github.com/rikvdh/go-tools/lib/todo"
)

const (
	port = ":50051"
)

// server is used to implement todo.TodoServer.
type server struct {
	db *bolt.DB
}

// itob returns an 8-byte big endian representation of v.
func itob(v uint64) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, uint64(v))
    return b
}

func mergeItem(dest, src *pb.TodoRequest) {
	// we never update: id, created, list
	// skip the title when it is empty
	if len(src.Title) > 0 {
		dest.Title = src.Title
	}
	// skip the description when it is empty
	if len(src.Description) > 0 {
		dest.Description = src.Description
	}
	if src.Priority != 0 {
		dest.Priority = src.Priority
	}
	dest.Done = src.Done
}

// CreateTodo creates a new Todo
func (s *server) Add(ctx context.Context, in *pb.TodoRequest) (*pb.TodoResponse, error) {
	reply := &pb.TodoResponse{Success: true}
	err := s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(in.List))
		if err != nil {
			return err
		}
		b := tx.Bucket([]byte(in.List))

		var item *pb.TodoRequest

		if in.Id > 0 {
			item = &pb.TodoRequest{}
			if err := json.Unmarshal(b.Get(itob(uint64(in.Id))), item); err != nil {
				return err
			}
			mergeItem(item, in)
		} else {
			// no ID present, add a new one
			item = in
			id, _ := b.NextSequence()
			item.Id = int32(id)
		}

		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		return b.Put(itob(uint64(item.Id)), data)
	})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// GetTodos returns all todos by given filter
func (s *server) List(filter *pb.TodoFilter, stream pb.Todo_ListServer) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(filter.List))
		if b == nil {
			fmt.Println("bucket is nil")
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%v, value=%s\n", k, v)
			td := &pb.TodoRequest{}
			if err := json.Unmarshal(v, td); err != nil {
				return err
			}
			if !td.Done || filter.All {
				if err := stream.Send(td); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Creates a new gRPC server
	s := grpc.NewServer()

	db, err := bolt.Open("todo.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	pb.RegisterTodoServer(s, &server{db: db})
	s.Serve(lis)
}
