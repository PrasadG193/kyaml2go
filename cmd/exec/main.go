package main

import (
	"bufio"
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
				return buildAndRun("create", c.Bool("cr"), c.String("client"), c.String("scheme"), c.String("apis"))
			},
		},
		{
			Name:  "update",
			Usage: "Generate code for updating a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun("update", c.Bool("cr"), c.String("client"), c.String("scheme"), c.String("apis"))
			},
		},
		{
			Name:  "get",
			Usage: "Generate code to get a resource object",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun("get", c.Bool("cr"), c.String("client"), c.String("scheme"), c.String("apis"))
			},
		},
		{
			Name:  "delete",
			Usage: "Generate code for deleting a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun("delete", c.Bool("cr"), c.String("client"), c.String("scheme"), c.String("apis"))
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

func buildAndRun(method string, isCR bool, client, scheme, api string) error {
	// Rebuild CLI with provided packages
	if isCR {
		// Generate register.go to register scheme as per the provided packages
		if out, err := execute("sh", []string{"-c", fmt.Sprintf("sed 's/PACKAGE/%s/g' ./pkg/generator/register_template.txt > ./pkg/generator/register.go", escape(scheme))}); err != nil {
			log.Printf("Failed to generate register.go %s. %v", out, err)
			return err
		}

		if out, err := execute("sh", []string{"-c", "make cli"}); err != nil {
			log.Printf("Failed build kyaml2go %s. %v", out, err)
			return err
		}
	}

	// Read input from the console
	var data string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading input:", err)
	}

	// Generate code
	args := os.Args[1:]
	kcli := fmt.Sprintf("%s/bin/kyaml2go_cli", os.Getenv("GOPATH"))
	c := exec.Command(kcli, args...)

	c.Stdin = strings.NewReader(data)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		log.Printf("Failed to exec kyaml2go binary. %v", err)
		return err
	}
	return nil
}
