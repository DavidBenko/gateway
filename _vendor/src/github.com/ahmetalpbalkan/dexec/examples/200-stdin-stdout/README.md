# 200-stdin-stdout

In this example, we start the `tr '[:lower:] [:upper:]'` process
in a container and we set its STDIN to a `strings.Reader` and set
its STDOUT to `os.Stdout`.

The `tr` invocation will convert passed stream to uppercase.

To run this example:

    $ docker pull busybox
    $ go run main.go
    HELLO WORLD
    FROM
    CONTAINER


## Migration from `os/exec`

```diff
> 	cl, _ := docker.NewClientFromEnv()
> 	d := dexec.Docker{cl}
> 
> 	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{
> 		Config: &docker.Config{Image: "busybox"}})
> 
< 	cmd := exec.Command("tr", "[:lower:]", "[:upper:]")
> 	cmd := d.Command(m, "tr", "[:lower:]", "[:upper:]")
```
