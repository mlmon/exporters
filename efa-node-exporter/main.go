package main

import (
	"context"
	"dagger.io/dagger"
	"log/slog"
	"os"
	"time"
)

const platform = "linux/amd64"

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

	builder := client.Git("https://github.com/aws-samples/awsome-distributed-training.git").
		Branch("main").
		Tree().
		Directory("4.validation_and_observability/3.efa-node-exporter").
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: "Dockerfile",
			Platform:   platform,
			BuildArgs: []dagger.BuildArg{
				{"GOLANG_VERSION", "1.23.2"},
				{"NODE_EXPORTER_VERSION", "v1.8.2"},
				{"PROCFS_EXPORTER_VERSION", "v0.14.0"},
				{"EFA_INSTALLER_VERSION", "1.30.0"},
			}})

	_, err = client.Container(dagger.ContainerOpts{Platform: platform}).
		From("public.ecr.aws/docker/library/ubuntu:20.04").
		WithFile("/bin/node_exporter", builder.File("/workspace/node_exporter/node_exporter")).
		WithEntrypoint([]string{"/bin/node_exporter"}).
		Publish(ctx, "nrfisher/efa-node-exporter:latest")
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}
