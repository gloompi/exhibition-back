package main

import (
	"fmt"
	"os"
	"strconv"
)

type config struct {
	port     int
	grpcPort int
}

func readConfig() config {
	portString := os.Getenv("PORT")
	grpcPortString := os.Getenv("GRPC_PORT")

	if portString == "" {
		portString = "9999"
	}

	if grpcPortString == "" {
		grpcPortString = "50051"
	}

	port, err := strconv.Atoi(portString)

	if err != nil {
		panic(fmt.Sprintf("Could not parse %s to int", portString))
	}

	grpcPort, err := strconv.Atoi(grpcPortString)

	if err != nil {
		panic(fmt.Sprintf("Could not parse %s to int", grpcPortString))
	}

	return config{port, grpcPort}
}
