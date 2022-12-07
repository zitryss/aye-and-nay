package dockertest

import (
	"net/url"
	"os"

	"github.com/ory/dockertest/v3"

	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New() docker {
	host, ok := os.LookupEnv("DOCKER_HOST")
	if !ok || host == "" {
		host = "tcp://localhost:2375"
	}
	u, err := url.Parse(host)
	if err != nil {
		err = errors.Wrap(err)
		log.Critical("dockertest:", err)
		os.Exit(1)
	}
	hostname := u.Hostname()
	pool, err := dockertest.NewPool("")
	if err != nil {
		err = errors.Wrap(err)
		log.Critical("could not connect to docker:", err)
		os.Exit(1)
	}
	resources := []*dockertest.Resource(nil)
	return docker{hostname, pool, resources}
}

type docker struct {
	host      string
	pool      *dockertest.Pool
	resources []*dockertest.Resource
}

func (d *docker) RunRedis(host *string, hPort *string) {
	repository := "redis"
	tag := "6-alpine"
	env := []string(nil)
	cmd := []string(nil)
	cPort := "6379/tcp"
	conf := func(port string) {
		*host = d.host
		*hPort = port
	}
	d.run(repository, tag, env, cmd, cPort, conf)
}

func (d *docker) RunImaginary(host *string, hPort *string) {
	repository := "h2non/imaginary"
	tag := "1"
	env := []string(nil)
	cmd := []string(nil)
	cPort := "9000/tcp"
	conf := func(port string) {
		*host = d.host
		*hPort = port
	}
	d.run(repository, tag, env, cmd, cPort, conf)
}

func (d *docker) RunMongo(host *string, hPort *string) {
	repository := "mongo"
	tag := "5"
	env := []string(nil)
	cmd := []string(nil)
	cPort := "27017/tcp"
	conf := func(port string) {
		*host = d.host
		*hPort = port
	}
	d.run(repository, tag, env, cmd, cPort, conf)
}

func (d *docker) RunMinio(host *string, hPort *string, accessKey string, secretKey string) {
	repository := "minio/minio"
	tag := "RELEASE.2021-11-24T23-19-33Z"
	env := []string{"MINIO_ACCESS_KEY=" + accessKey, "MINIO_SECRET_KEY=" + secretKey}
	cmd := []string{"server", "/data"}
	cPort := "9000/tcp"
	conf := func(port string) {
		*host = d.host
		*hPort = port
	}
	d.run(repository, tag, env, cmd, cPort, conf)
}

func (d *docker) run(repository string, tag string, env []string, cmd []string, containerPort string, conf func(string)) {
	resource, err := d.pool.RunWithOptions(&dockertest.RunOptions{Repository: repository, Tag: tag, Env: env, Cmd: cmd})
	if err != nil {
		err = errors.Wrap(err)
		log.Critical("could not start resource:", err)
		os.Exit(1)
	}
	hostPort := resource.GetPort(containerPort)
	conf(hostPort)
	d.resources = append(d.resources, resource)
}

func (d *docker) Purge() {
	for _, r := range d.resources {
		err := d.pool.Purge(r)
		if err != nil {
			err = errors.Wrap(err)
			log.Critical("could not purge resource:", err)
			os.Exit(1)
		}
	}
}
