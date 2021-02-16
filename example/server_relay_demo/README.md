server console

```
./server_relay_demo
```

publisher console
```
curl -L https://test-videos.co.uk/vids/bigbuckbunny/mp4/h264/1080/Big_Buck_Bunny_1080_10s_1MB.mp4 -o movie.mp4
ffmpeg -re -stream_loop -1 -i movie.mp4 -acodec copy -vcodec copy -f flv rtmp://localhost/appname/stream
```

subscriber console
```

```
