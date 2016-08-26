# 400-exit-codes

Just like `os/exec` package, `dexec` also returns an `ExitError`
type that you can collect valuable information such as:

* `ExitCode int`
* `Stderr   []byte` if process was started using `Output()` method.

In this example, we start a program that is supposed to fail:

    /bin/sh -c "exit 255;"

To run this example:

    $ docker pull busybox
    $ go run main.go
    exit code=255

## Migration from `os/exec`

```diff
> 	cl, _ := docker.NewClientFromEnv()
> 	d := dexec.Docker{cl}
> 
> 	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{
> 		Config: &docker.Config{Image: "busybox"}})
> 
< 	cmd := exec.Command("sh", "-c", "exit 255;")
> 	cmd := d.Command(m, "sh", "-c", "exit 255;")
< 	if ee, ok := err.(*exec.ExitError); ok {
< 		fmt.Printf("exit code=%d\n", ee.Sys().(syscall.WaitStatus).ExitStatus())
> 	if ee, ok := err.(*dexec.ExitError); ok {
> 		fmt.Printf("exit code=%d\n", ee.ExitCode)
```
