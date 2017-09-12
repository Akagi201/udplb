package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Akagi201/udplb/config"
	"github.com/Akagi201/udplb/log"
	"github.com/Akagi201/udplb/server"
	"github.com/Akagi201/udplb/upstream"
)

func main() {
	var upstreams []*upstream.Upstream
	var servers []*server.Server

	log.Info("Loading configuration file")
	settings, err := config.Load(config.Opts.Conf)
	if err != nil {
		log.Errorf("Can't read configuration file: %s\n", err.Error())
	}

	// Parse upstreams first
	// TODO: Kill AutoUpdate gracefully
	upstreams = make([]*upstream.Upstream, len(settings.Upstreams))
	for i := range settings.Upstreams {
		ups := upstream.MustNewUpstream(&settings.Upstreams[i])

		if ups.IsDynamic {
			go upstream.AutoUpdatePeer(ups, 300)
		}

		upstreams[i] = ups
	}

	// Then parse servers
	servers = make([]*server.Server, len(settings.Servers))
	for i := range settings.Servers {
		server, err := server.NewServer(&settings.Servers[i], upstreams)

		if err != nil {
			log.Error(err)
			return
		}

		log.Infof("Starting server on port: %d\n", server.Config.Port)
		server.MustStart()
		servers[i] = server
	}

	// Wait for a termination signal
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	// Stop the service gracefully.
	for _, s := range servers {
		s.Stop()
	}
}
