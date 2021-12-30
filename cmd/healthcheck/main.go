package main

import (
	"flag"
	"io"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

func main() {
	conf := ""
	flag.StringVar(&conf, "config", "./config.yml", "relative filepath to a config file")
	flag.Parse()
	viper.SetConfigFile(conf)
	err := viper.ReadInConfig()
	if err != nil {
		os.Exit(1)
	}
	port := viper.GetString("server.port")
	req, err := http.NewRequest(http.MethodGet, "http://localhost:"+port+"/api/health/", http.NoBody)
	if err != nil {
		os.Exit(1)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		os.Exit(1)
	}
	if resp.StatusCode != 200 {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		os.Exit(1)
	}
	_, err = io.Copy(io.Discard, resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		os.Exit(1)
	}
	err = resp.Body.Close()
	if err != nil {
		os.Exit(1)
	}
}
