package server

import (
	"github.com/zfiona/server-base/conf"
	"github.com/zfiona/server-base/log"
	"github.com/zfiona/server-base/module"
	"os"
	"os/signal"
)

const version = "1.0.0"

func Run(mods ...module.Module) {
	// logger
	if conf.LogLevel != "" {
		logger, err := log.New(conf.LogLevel, conf.LogPath, conf.LogFlag)
		if err != nil {
			panic(err)
		}
		log.Export(logger)
		defer logger.Close()
	}

	log.Release("server %v starting up", version)

	// module
	for i := 0; i < len(mods); i++ {
		module.Register(mods[i])
	}
	module.Init()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	module.Destroy()
	log.Release("server closing down (signal: %v)", sig)
}
