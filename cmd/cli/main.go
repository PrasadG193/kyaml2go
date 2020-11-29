package main

import (
	"bufio"
	"fmt"
	"github.com/urfave/cli"
	//"io/ioutil"
	"log"
	"os"

	"github.com/PrasadG193/kyaml2go/cmd/option"
	gen "github.com/PrasadG193/kyaml2go/pkg/generator"
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
				return generate(gen.MethodCreate, c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("apis"))
			},
		},
		{
			Name:  "update",
			Usage: "Generate code for updating a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return generate(gen.MethodUpdate, c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("apis"))
				//return generate(c.String("file"), gen.MethodUpdate)
			},
		},
		{
			Name:  "get",
			Usage: "Generate code to get a resource object",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return generate(gen.MethodGet, c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("apis"))
			},
		},
		{
			Name:  "delete",
			Usage: "Generate code for deleting a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return generate(gen.MethodDelete, c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("apis"))
			},
		},
	}

	// Run app
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func generate(method gen.KubeMethod, isCR, isNamespaced bool, client, api string) error {
	// Read input from the console
	var data string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Error while reading input:", err)
	}

	//b, err := ioutil.ReadFile(path)
	//if err != nil {
	//	return cli.NewExitError(fmt.Errorf("error: the path %s does not exist", path), 1)
	//}
	//gen := gen.New(b, method, true, true, "k8s.io/sample-controller/pkg/generated/clientset/versioned", "k8s.io/sample-controller/pkg/apis/samplecontroller")
	//fmt.Printf("DATA::\n%s\n", data)
	gen := gen.New([]byte(data), method, isCR, isNamespaced, client, api)
	code, err := gen.Generate()
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	fmt.Println(code)
	return nil
}
