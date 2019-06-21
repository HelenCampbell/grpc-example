package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/scottyw/grpc-example/D.adding-endpoint/factory"
	"google.golang.org/grpc"
)

func startGRPC() {
	lis, err := net.Listen("tcp", "localhost:5566")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	factory.RegisterBoxFactoryServer(grpcServer, &factoryServer{})
	log.Println("gRPC server ready ...")
	grpcServer.Serve(lis)
}

func startHTTP() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Connect to the GRPC server
	conn, err := grpc.Dial("localhost:5566", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	// Register grpc-gateway
	rmux := runtime.NewServeMux()
	client := factory.NewBoxFactoryClient(conn)
	err = factory.RegisterBoxFactoryHandlerClient(ctx, rmux, client)
	if err != nil {
		log.Fatal(err)
	}

	// Serve the swagger, swagger-ui and grpc-gateway REST bindings on 8080
	mux := http.NewServeMux()
	mux.Handle("/v1/", rmux)
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("www"))))
	log.Println("REST server ready...")
	log.Println("Serving Swagger at: http://localhost:8080/swagger-ui/")
	err = http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	go startGRPC()

	go startHTTP()

	// Block forever
	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()

}
