package control

import (
	"fmt"
	"sort"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	"github.com/rancher/os/compose"
	"github.com/rancher/os/config"
	"github.com/rancher/os/util/network"
)

func engineSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "switch",
			Usage:  "switch Docker engine without a reboot",
			Action: engineSwitch,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "do not prompt for input",
				},
				cli.BoolFlag{
					Name:  "no-pull",
					Usage: "don't pull console image",
				},
			},
		},
		{
			Name:   "enable",
			Usage:  "set Docker engine to be switched on next reboot",
			Action: engineEnable,
		},
		{
			Name:   "list",
			Usage:  "list available Docker engines",
			Action: engineList,
		},
	}
}

func engineSwitch(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one Docker engine to switch to")
	}
	newEngine := c.Args()[0]

	cfg := config.LoadConfig()

	project, err := compose.GetProject(cfg, true, false)
	if err != nil {
		log.Fatal(err)
	}

	if err = project.Stop(context.Background(), 10, "docker"); err != nil {
		log.Fatal(err)
	}

	project.ServiceConfigs.Add("docker", &composeConfig.ServiceConfig{})

	if err = compose.LoadService(project, cfg, true, newEngine); err != nil {
		log.Fatal(err)
	}

	if err = project.Up(context.Background(), options.Up{}, "docker"); err != nil {
		log.Fatal(err)
	}

	if err := config.Set("rancher.docker.engine", newEngine); err != nil {
		log.Errorf("Failed to update rancher.docker.engine: %v", err)
	}

	return nil
}

func engineEnable(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one Docker engine to enable")
	}
	newEngine := c.Args()[0]

	cfg := config.LoadConfig()

	if err := compose.StageServices(cfg, newEngine); err != nil {
		return err
	}

	if err := config.Set("rancher.docker.engine", newEngine); err != nil {
		log.Errorf("Failed to update 'rancher.docker.engine': %v", err)
	}

	return nil
}

func engineList(c *cli.Context) error {
	cfg := config.LoadConfig()

	engines, err := network.GetEngines(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		return err
	}
	sort.Strings(engines)

	for _, engine := range engines {
		fmt.Println(engine)
	}

	return nil
}
