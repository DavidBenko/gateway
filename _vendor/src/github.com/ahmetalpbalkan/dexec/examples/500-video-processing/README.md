# 500-video-processing

In this example, we will extract audio (mp3) from a YouTube video
using `youtubbe-dl` and `ffmpeg`.

However instead of downloading the entire video to our computer and
keeping our CPU busy, we can offload this work to cloud using `dexec`: 

    youtube-dl [url] -o - | ffmpeg -i pipe:0 -f mp3 pipe:1

Then, we save the STDOUT of this command to `music.mp3`.

> Sorry for teaching you how to pirate music, I couldn’t find a better
> example.

To run this example:

    $ docker pull vimagick/youtube-dl
    $ go run main.go
    (play music.mp3)

## Migration from `os/exec`

Again, just 5 lines (after 4 examples, you’re not surprised, right?):

```diff
> 	cl, _ := docker.NewClientFromEnv()
> 	d := dexec.Docker{cl}
> 
> 	m, _ := dexec.ByCreatingContainer(docker.CreateContainerOptions{
> 		Config: &docker.Config{Image: "vimagick/youtube-dl"}})
< 	cmd := exec.Command("sh", "-c", fmt.Sprintf("youtube-dl %s -o - | ffmpeg -i pipe:0 -f mp3 pipe:1", url))
> 	cmd := d.Command(m, "sh", "-c", fmt.Sprintf("youtube-dl %s -o - | ffmpeg -i pipe:0 -f mp3 pipe:1", url))
```
