package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/PrasadG193/kyaml2go/cmd/option"
)

var flags = option.Flags

func main() {
	app := cli.NewApp()
	app.Name = "kyaml2go"
	app.Usage = "Generate go code to manage Kubernetes resources using client-go sdks"

	app.Commands = []cli.Command{
		{
			Name:  "create",
			Usage: "Generate code for creating a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun(c.String("file"), "create", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("schema"), c.String("apis"))
			},
		},
		{
			Name:  "update",
			Usage: "Generate code for updating a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun(c.String("file"), "update", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("schema"), c.String("apis"))
			},
		},
		{
			Name:  "get",
			Usage: "Generate code to get a resource object",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun(c.String("file"), "get", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("schema"), c.String("apis"))
			},
		},
		{
			Name:  "delete",
			Usage: "Generate code for deleting a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun(c.String("file"), "delete", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("schema"), c.String("apis"))
			},
		},
	}

	// Run app
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func execute(cmd string, args []string) (string, error) {
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	return string(out), err
}

func escape(p string) string {
	return strings.ReplaceAll(p, "/", "\\/")
}

func buildAndRun(path string, method string, isCR, isNamespaced bool, client, schema, api string) error {
	if isCR {
		if out, err := execute("sh", []string{"-c", fmt.Sprintf("sed 's/PACKAGE/%s/g' ./pkg/generator/register_template.txt > ./pkg/generator/register.go", escape(schema))}); err != nil {
			log.Printf("Failed to generate register.go %s. %v", out, err)
			return err
		}
	}

	if out, err := execute("sh", []string{"-c", "make cli"}); err != nil {
		log.Printf("Failed build kyaml2go %s. %v", out, err)
		return err
	}

	//if out, err := execute("go", []string{"build", "./cmd/cli"}); err != nil {
	//	log.Printf("Failed build kyaml2go %s. %v", out, err)
	//	return err
	//}

	out, err := execute(fmt.Sprintf("%s/bin/kyaml2go_cli", os.Getenv("GOPATH")), os.Args[1:])
	if err != nil {
		log.Printf("Failed to exec kyaml2go binary %s. %v", out, err)
		return err
	}
	fmt.Println(out)
	return nil
}
