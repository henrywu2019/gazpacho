package main

import (
    "fmt"

    cfg "github.com/vyskocilm/gazpacho/g/cfg"
)

type Config struct {
    Name string `yaml:"name"`
    Verbose bool `yaml:"verbose"`
}

func main() {

    var conf Config
    err := cfg.Load(&conf)
    if err != nil {
        panic(err)
    }
    fmt.Printf("conf=%s\n", conf)
}
