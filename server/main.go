/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	pb "github.com/roulette_grpc/roulette"
	"github.com/roulette_grpc/server/lib"

	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGameServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) Play(ctx context.Context, in *pb.GameRequest) (*pb.GameReply, error) {
	options, err := readGameOptions(in.GetName())
	if err != nil {
		return nil, fmt.Errorf("EmptyGameOptions: %w", err)
	}
	optionBytes, _ := json.MarshalIndent(options, "", "  ")
	log.Printf("Received: %s", string(optionBytes))

	games := lib.NewGames(options)
	if games == nil {
		return nil, fmt.Errorf("EmptyGames: %w", err)
	}

	stats := GetStats(*games)
	return &pb.GameReply{Message: stats}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGameServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func GetStats(games lib.Games) string {
	results, resultsEach, err := games.Play()
	var output string
	if err != nil {
		output += fmt.Sprintf("%s\n", err)
	} else {
		output += "------------------------------------\n"
		output += "p(<wage>) = <prob>\n"
		output += "------------------------------------\n"
		output += fmt.Sprintf("%s\n", results)
		output += "------------------------------------\n"
		output += "Descriptive Stats (Summary)\n"
		output += "------------------------------------\n"
		output += statsResults(*results)
		output += "------------------------------------\n"
		output += "Descriptive Stats (Individual)\n"
		output += "------------------------------------\n"
		for i, eachResult := range resultsEach {
			output += statsResults(*eachResult, i)
		}
	}
	return output
}

func statsResults(results lib.Results, index ...int) string {
	var output string
	if len(index) > 0 {
		output += fmt.Sprintf("[%d] ", index[0])
	}

	stats, err := results.Stats()
	if err != nil {
		output += fmt.Sprintf("%s\n", err)
	} else {
		output += fmt.Sprintf("%s\n", stats)
	}
	return output
}

func readGameOptions(filename string) ([]lib.GameOptions, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("InvalidFilename: %w", err)
	}
	bytes, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("InvalidFile: %w", err)
	}
	var options []lib.GameOptions
	if err := json.Unmarshal(bytes, &options); err != nil {
		return nil, fmt.Errorf("InvalidUnmarshal: %w", err)
	}
	return options, nil
}
