## Server relay example

A minimum example to relay RTMP streams on local. Streams will be published and can be subscribed per publish name.

### server console

```
make
./server_relay_demo
```

### publisher console

```
curl -L https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_1MB.mp4 -o movie.mp4
ffmpeg -re -stream_loop -1 -i movie.mp4 -acodec copy -vcodec copy -f flv rtmp://localhost/appname/stream
```

### subscriber console

```
ffplay rtmp://localhost/appname/stream
```
