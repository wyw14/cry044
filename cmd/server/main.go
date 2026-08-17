package main

import (
	"github.com/wyw14/cry044/internal/config"
	"github.com/wyw14/cry044/internal/platform"
)

func main() { platform.Serve(config.Load()) }
