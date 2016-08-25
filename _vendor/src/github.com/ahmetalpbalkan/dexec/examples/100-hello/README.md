# 100-hello-world

Welcome to `dexec`! First of all, thank you for your interest in
learning more.

In this brief example, the command: `echo "I am running inside a container!"`
and collect its output to a `[]byte` to print to screen.

To run this example:

    $ docker pull busybox
    $ go run main.go
    I am running inside a container!

## Migration from `os/exec`

```diff

> 	cl, _ := docker.NewClientFromEnv()
> 	d := dexec.Docker{cl}
> 
> 	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{
> 		Config: &docker.Config{Image: "busybox"}})
< 	cmd := exec.Command("echo", `I am running inside a container!`)
> 	cmd := d.Command(m, "echo", `I am running inside a container!`)
```
