package main

import (
	"github.com/ahmetalpbalkan/go-dexec"
	"github.com/fsouza/go-dockerclient"
	"os"
	"strings"
)

func main() {
	input := `Hello world
from
container`

	cl, _ := docker.NewClientFromEnv()
	d := dexec.Docker{cl}

	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{
		Config: &docker.Config{Image: "busybox"}})

	cmd := d.Command(m, "tr", "[:lower:]", "[:upper:]")
	cmd.Stdin = strings.NewReader(input) // <--
	cmd.Stdout = os.Stdout               // <--

	if err := cmd.Run(); err != nil {
		panic(err)
	}
	// Output:
	//   HELLO WORLD
	//   FROM
	//   CONTAINER
}
