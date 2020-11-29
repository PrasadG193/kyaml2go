package main

import (
	"bufio"
	//"bytes"
	"fmt"
	"github.com/urfave/cli"
	//"io"
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
				return buildAndRun("create", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("scheme"), c.String("apis"))
			},
		},
		{
			Name:  "update",
			Usage: "Generate code for updating a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun("update", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("scheme"), c.String("apis"))
			},
		},
		{
			Name:  "get",
			Usage: "Generate code to get a resource object",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun("get", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("scheme"), c.String("apis"))
			},
		},
		{
			Name:  "delete",
			Usage: "Generate code for deleting a resource",
			Flags: flags,
			Action: func(c *cli.Context) error {
				return buildAndRun("delete", c.Bool("cr"), c.Bool("namespaced"), c.String("client"), c.String("scheme"), c.String("apis"))
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
	log.Println("command: ", cmd, args)
	c := exec.Command(cmd, args...)
	out, err := c.CombinedOutput()
	return string(out), err
}

func escape(p string) string {
	return strings.ReplaceAll(p, "/", "\\/")
}

func buildAndRun(method string, isCR, isNamespaced bool, client, scheme, api string) error {
	if isCR {
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

	//fmt.Printf("DATA EXEC::\n%s\n", data)
	// hack
	//path := "/tmp/manifest.yaml"
	//file, err := os.OpenFile(path, os.O_RDWR, 0644)
	//if err != nil {
	//	// create file if not exists
	//	if os.IsNotExist(err) {
	//		file, err = os.Create(path)
	//		if err != nil {
	//			log.Println(err)
	//			return err
	//		}
	//	}
	//}
	//defer file.Close()
	//defer os.Remove(path)
	//_, err = file.WriteString(string(data))
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}

	//if out, err := execute("go", []string{"build", "./cmd/cli"}); err != nil {
	//	log.Printf("Failed build kyaml2go %s. %v", out, err)
	//	return err
	//}
	//out1, err := execute("echo", []string{path})
	//if err != nil {
	//	log.Printf("Failed to echo %s. %v", out, err)
	//	return err
	//}
	//log.Printf("FILE CONTENT %s\b", string(out1))

	//args := append(os.Args[1:], []string{"<", fmt.Sprintf("<(echo '%s')", data)}...)
	args := os.Args[1:]
	//args := append(os.Args[1:], []string{"<", path}...)
	kcli := fmt.Sprintf("%s/bin/kyaml2go_cli", os.Getenv("GOPATH"))
	//c := exec.Command("sh", append([]string{"-c", kcli}, args...)...)
	c := exec.Command(kcli, args...)

	c.Stdin = strings.NewReader(data)
	//var out bytes.Buffer
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()

	//stdin, err := c.StdinPipe()
	//if err != nil {
	//	return err
	//}

	////go func() {
	//defer stdin.Close()
	//io.WriteString(stdin, data)
	////}()

	////c.Stdin = strings.NewReader(data)
	//out, err := c.CombinedOutput()
	//out, err := execute("sh", append([]string{"-c", kcli}, args...))
	if err != nil {
		log.Printf("Failed to exec kyaml2go binary. %v", err)
		//log.Printf("Failed to exec kyaml2go binary %s. %v", out.String(), err)
		return err
	}
	//fmt.Println(out.String())
	return nil
}
