# 300-pipes (Stream Processing)

Let’s make the previous example a bit more stream-like.

This time we use pipes to provide input as stream and collect
results as they are processed.

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
