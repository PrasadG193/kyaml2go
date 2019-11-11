package main

import (
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"

	gen "github.com/PrasadG193/kgoclient-gen/pkg/generator"
)

func main() {

	app := cli.NewApp()
	app.Name = "kgoclientgen"
	app.Usage = "Generate go code to manage Kubernetes resources using client-go sdks"

	app.Commands = []cli.Command{
		{
			Name:  "create",
			Usage: "Generate code for creating a resource",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "file, f",
					Usage:    "K8s resource spec yaml file",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				return generate(c.String("file"), gen.MethodCreate)
			},
		},
		{
			Name:  "update",
			Usage: "Generate code for updating a resource",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "file, f",
					Usage:    "K8s resource spec yaml file",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				return generate(c.String("file"), gen.MethodUpdate)
			},
		},
		{
			Name:  "delete",
			Usage: "Generate code for deleting a resource",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:     "file, f",
					Usage:    "K8s resource spec yaml file",
					Required: true,
				},
			},
			Action: func(c *cli.Context) error {
				return generate(c.String("file"), gen.MethodDelete)
			},
		},
	}

	// Run app
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func generate(path string, method gen.KubeMethod) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("error: the path %s does not exist", path), 1)
	}
	gen := gen.New(b, method)
	code, err := gen.Generate()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(code)
	return nil
}
