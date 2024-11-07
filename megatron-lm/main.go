package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"log/slog"
	"os"
	"time"
)

const (
	platform      = "linux/amd64"
	awsEfaVersion = "1.34.0"
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
		Directory("3.test_cases/1.megatron-lm").
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: "0.distributed-training.Dockerfile",
			Platform:   platform,
			BuildArgs: []dagger.BuildArg{
				{"EFA_INSTALLER_VERSION", awsEfaVersion},
			}}).
		WithAnnotation("org.opencontainers.image.title", "efa-node-exporter").
		WithAnnotation("org.opencontainers.image.url", "https://github.com/mlmon/exporters/").
		WithAnnotation("org.opencontainers.image.version", fmt.Sprintf("aws-efa:%s,", awsEfaVersion)).
		Publish(ctx, "nrfisher/megatron-ml:latest")
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}
