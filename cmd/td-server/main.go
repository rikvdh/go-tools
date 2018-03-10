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
var (
	bucketName = []byte("rikvdh")
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

// CreateTodo creates a new Todo
func (s *server) Add(ctx context.Context, in *pb.TodoRequest) (*pb.TodoResponse, error) {
	reply := &pb.TodoResponse{Success: true}
	err := s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err == nil {
			b := tx.Bucket(bucketName)
			if in.Id == 0 {
				// no ID present, add a new one
				id, _ := b.NextSequence()
				reply.Id = int32(id)
				in.Id = reply.Id
				data, err := json.Marshal(in)
				if err != nil {
					return err
				}

				return b.Put(itob(id), data)
			} else {
				td := &pb.TodoRequest{}
				id := itob(uint64(in.Id))
				if err := json.Unmarshal(b.Get(id), td); err != nil {
					return err
				}
				if in.Done {
					td.Done = in.Done
				}
				data, err := json.Marshal(td)
				if err != nil {
					return err
				}
				return b.Put(id, data)
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// GetTodos returns all todos by given filter
func (s *server) List(filter *pb.TodoFilter, stream pb.Todo_ListServer) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
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
			if err := stream.Send(td); err != nil {
				return err
			}
		}

		return nil
	})
/*	for _, todo := range s.savedTodos {
			if !strings.Contains(todo.Title, filter.Text) {
				continue
			}
		if filter.Text != "" {
		}
	}
	return nil*/
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
