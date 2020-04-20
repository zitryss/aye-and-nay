package dockertest

import (
	"net/url"
	"os"

	"github.com/ory/dockertest/v3"
	"github.com/spf13/viper"

	"github.com/zitryss/aye-and-nay/pkg/env"
	"github.com/zitryss/aye-and-nay/pkg/errors"
	"github.com/zitryss/aye-and-nay/pkg/log"
)

func New() docker {
	host, err := env.Lookup("DOCKER_HOST")
	if err != nil {
		log.Critical("dockertest: ", err)
		os.Exit(1)
	}
	u, err := url.Parse(host)
	if err != nil {
		err = errors.Wrap(err)
		log.Critical("dockertest: ", err)
		os.Exit(1)
	}
	hostname := u.Hostname()
	pool, err := dockertest.NewPool("")
	if err != nil {
		err = errors.Wrap(err)
		log.Critical("could not connect to docker:", err)
		os.Exit(1)
	}
	return docker{hostname, pool, nil}
}

type docker struct {
	host      string
	pool      *dockertest.Pool
	resources []*dockertest.Resource
}

func (d *docker) RunMinio() {
	repository := "minio/minio"
	tag := "RELEASE.2020-04-15T19-42-18Z"
	accessKey := viper.GetString("minio.accessKey")
	secretKey := viper.GetString("minio.secretKey")
	env := []string{"MINIO_ACCESS_KEY=" + accessKey, "MINIO_SECRET_KEY=" + secretKey}
	cmd := []string{"server", "/data"}
	port := "9000/tcp"
	conf := func(port string) {
		viper.Set("minio.host", d.host)
		viper.Set("minio.port", port)
	}
	d.run(repository, tag, env, cmd, port, conf)
}

func (d *docker) RunMongo() {
	repository := "mongo"
	tag := "4"
	env := []string(nil)
	cmd := []string(nil)
	port := "27017/tcp"
	conf := func(port string) {
		viper.Set("mongo.host", d.host)
		viper.Set("mongo.port", port)
	}
	d.run(repository, tag, env, cmd, port, conf)
}

func (d *docker) RunRedis() {
	repository := "redis"
	tag := "5-alpine"
	env := []string(nil)
	cmd := []string(nil)
	port := "6379/tcp"
	conf := func(port string) {
		viper.Set("redis.host", d.host)
		viper.Set("redis.port", port)
	}
	d.run(repository, tag, env, cmd, port, conf)
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
