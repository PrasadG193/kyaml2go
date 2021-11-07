package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/PrasadG193/kyaml2go/cmd/option"
	gen "github.com/PrasadG193/kyaml2go/pkg/generator"
	"github.com/PrasadG193/kyaml2go/pkg/types"
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
				return generate(types.MethodCreate, c.Bool("cr"), c.Bool("namespaced"), c.Bool("dynamic"), c.String("client"), c.String("apis"))
			},
		},
		{
			Name:  "update",
			Usage: "Generate code for updating a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return generate(types.MethodUpdate, c.Bool("cr"), c.Bool("namespaced"), c.Bool("dynamic"), c.String("client"), c.String("apis"))
			},
		},
		{
			Name:  "get",
			Usage: "Generate code to get a resource object",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return generate(types.MethodGet, c.Bool("cr"), c.Bool("namespaced"), c.Bool("dynamic"), c.String("client"), c.String("apis"))
			},
		},
		{
			Name:  "delete",
			Usage: "Generate code for deleting a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return generate(types.MethodDelete, c.Bool("cr"), c.Bool("namespaced"), c.Bool("dynamic"), c.String("client"), c.String("apis"))
			},
		},
	}

	// Run app
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func generate(method types.KubeMethod, isCR, namespaced, isDynamic bool, clientpkg, api string) error {
	// Read input from the console
	var data string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data += scanner.Text() + "\n"
	}

	// Set default client and api packages if not passed
	if isCR {
		if clientpkg == "" {
			clientpkg = "github.com/PATH/TO/TYPED/GENERATED/CLIENTSET/versioned"
		}
		if api == "" {
			api = "github.com/PATH/TO/APIS/PACKAGE/resource"
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading input:", err)
	}

	gen := gen.New([]byte(data), method, isCR, namespaced, isDynamic, clientpkg, api)
	code, err := gen.Generate()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(code)
	return nil
}
