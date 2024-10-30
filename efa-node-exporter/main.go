package main

import (
	"context"
	"dagger.io/dagger"
	"log/slog"
	"os"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	duration := time.Now().Add(time.Hour * 2)
	ctx, cancel := context.WithDeadline(context.Background(), duration)
	defer cancel()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stderr))
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
	defer client.Close()

	_, err = client.Git("https://github.com/aws-samples/awsome-distributed-training.git").
		Branch("main").
		Tree().
		Directory("4.validation_and_observability/3.efa-node-exporter").
		DockerBuild(dagger.DirectoryDockerBuildOpts{Dockerfile: "Dockerfile", Platform: "linux/amd64"}).
		Publish(ctx, "nrfisher/efa-node-exporter:latest")
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}
