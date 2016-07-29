# 600-parallel-compute

In this example, we distibute a Compute and Network-heavy workload to
a Swarm cluster, by again, changing only a few lines of code from `os/exec`.

The program reads URLs from `files.txt` and starts containers in parallel
to download and compute MD5 checksum of the file using `wget -O- [url] | md5sum`.

To run this example:

1. Create a swarm cluster, edit its IP address and cert path in `main.go`.
2. Set your env vars to point to Swarm master:
3. Run:

```sh
$ docker pull busybox
$ go run main.go
```

### `os/exec` vs `dexec`

| Mode  | Execution Time | Bytes Downloaded | 
| ------| -------------- | ---------------- |
| Locally on a single node (`os/exec`)    |  `1m20s` | `0.8 GB` |
| 5-node Docker Swarm Cluster (`dexec`) on cloud | `10s` | `396 bytes` |


### Migration from `os/exec`

Good ol’ 5-lines:

```diff
> 	cl, _ := docker.NewClientFromEnv()
> 	d = dexec.Docker{cl}
> 	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{
> 		Config: &docker.Config{Image: "busybox"}})
< 	cmd := exec.Command("sh", "-c", fmt.Sprintf("wget -qO- %s | md5sum", url))
> 	cmd := d.Command(m, "sh", "-c", fmt.Sprintf("wget -qO- %s | md5sum", url))
```
