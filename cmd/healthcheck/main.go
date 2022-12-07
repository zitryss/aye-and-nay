package main

import (
	"flag"
	"os"

	"github.com/zitryss/aye-and-nay/internal/client"
	"github.com/zitryss/aye-and-nay/internal/config"
)

func main() {
	path := ""
	flag.StringVar(&path, "config", "./config.env", "filepath to a config file")
	flag.Parse()
	conf, err := config.New(path)
	if err != nil {
		os.Exit(1)
	}
	api := "http://localhost:" + conf.Server.Port
	timeout := conf.Server.WriteTimeout
	c, err := client.New(api, timeout)
	if err != nil {
		os.Exit(1)
	}
	err = c.Health()
	if err != nil {
		os.Exit(1)
	}
}
