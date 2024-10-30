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
	platform            = "linux/amd64"
	awsEfaVersion       = "1.30.0"
	nodeExporterVersion = "v1.8.2"
	procFsVersion       = "v0.14.0"
	goVersion           = "1.23.2"
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

	builder := client.Git("https://github.com/aws-samples/awsome-distributed-training.git").
		Branch("main").
		Tree().
		Directory("4.validation_and_observability/3.efa-node-exporter").
		DockerBuild(dagger.DirectoryDockerBuildOpts{
			Dockerfile: "Dockerfile",
			Platform:   platform,
			BuildArgs: []dagger.BuildArg{
				{"GOLANG_VERSION", goVersion},
				{"NODE_EXPORTER_VERSION", nodeExporterVersion},
				{"PROCFS_EXPORTER_VERSION", procFsVersion},
				{"EFA_INSTALLER_VERSION", awsEfaVersion},
			}})

	_, err = client.Container(dagger.ContainerOpts{Platform: platform}).
		From("public.ecr.aws/docker/library/ubuntu:20.04").
		WithFile("/bin/node_exporter", builder.File("/workspace/node_exporter/node_exporter")).
		WithEntrypoint([]string{"/bin/node_exporter"}).
		WithAnnotation("org.opencontainers.image.title", "efa-node-exporter").
		WithAnnotation("org.opencontainers.image.url", "https://github.com/mlmon/exporters/").
		WithAnnotation("org.opencontainers.image.version", fmt.Sprintf("go:%s;node-exporter:%s;aws-efa:%s,", goVersion, nodeExporterVersion, awsEfaVersion)).
		Publish(ctx, "nrfisher/efa-node-exporter:latest")
	if err != nil {
		logger.Error("%v", err)
		os.Exit(1)
	}
}
