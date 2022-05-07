package main

import (
	"flag"
	"log"
	"os"
	"sync"

	"github.com/stenehall/gosh/internal/config"
	"github.com/stenehall/gosh/internal/favii"
	"github.com/stenehall/gosh/internal/healthcheck"
	"github.com/stenehall/gosh/internal/web"
)

var version string

func main() {
	env := os.Getenv("APP_ENV")
	log.Printf("Starting %s %s...", env, version)

	configFlag := flag.String("config", "./config.yml", "Yaml config to use")
	healthCheckFlag := flag.Bool("health", false, "Check if we're healthy")
	flag.Parse()

	conf, err := config.New(*configFlag)

	// Poor man health-check
	if *healthCheckFlag {
		healthcheck.Healthcheck(conf.Data.Port)
	}
	if err != nil {
		panic(err)
	}

	log.Printf("Loaded configs, found %v sets containing %v sites\n", conf.CountSets(), conf.CountSites())

	fav := favii.Setup()

	var waitGroup sync.WaitGroup
	for setIndex, set := range conf.Data.Sets {
		for siteIndex, site := range set.Sites {
			waitGroup.Add(1)

			go func(setIndex int, siteIndex int, site favii.Site) {
				defer waitGroup.Done()

				updatedSite, err := fav.UpdateSite(site)
				if err != nil {
					log.Println(err)
				}
				if err == nil {
					conf.Data.Sets[setIndex].Sites[siteIndex] = updatedSite
				}
			}(setIndex, siteIndex, site)
		}
	}
	waitGroup.Wait()

	if env == "production" {
		if err := conf.SaveConfig(); err != nil {
			log.Printf("couldn't save config %v", err)
		}
	}

	if err := web.Ginning(conf.Data); err != nil {
		log.Printf("error from Gin %v", err)
	}
}
