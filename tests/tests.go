package tests

import (
	"log"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
)

func init() {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatal(err)
	}

	opts := &dockertest.RunOptions{
		Hostname:     "mongodb",
		Name:         "mongodb",
		Repository:   "mongo",
		Tag:          "4.4.6",
		ExposedPorts: []string{"27017"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"27017": {{HostIP: "0.0.0.0", HostPort: "27017"}},
		},
	}
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		config.AutoRemove = true
	})
	if err != nil {
		log.Fatal(err)
	}

	resource.Expire(10)
}
