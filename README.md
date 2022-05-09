# torrenttp
Simple REST API controlled BitTorrent client

## Docker
```
docker run -d \
--name torrenttp \
-p 1010:1010 \
-v ~/tttpdl:/dl \
--restart unless-stopped \
glbl/torrenttp:latest
```

## Compiling
```
go mod download
go build -ldflags="-extldflags -static -w -s" -tags=nosqlite
```

## Usage
`torrenttp [-dir DOWNLOADDIR] [-port PORT] [-noup]`

## API

### Add torrent
`POST /api/addtorrent`

For magnet link
```
{
    "magnet": "MAGNET_LINK"
}
```

For manual infohash, display name, and trackers
```
{
    "infohash": "INFOHASH",
    "displayname": "DISPLAY_NAME",
    "trackers": ["TRACKER_1", "TRACKER_2"]
}
```

### Selecting a file for download
`POST /api/selectfile`

```
{
    "infohash": "INFOHASH",
    "allfiles": false,
    "files": ["FILE_1", "FILE_2"]
}
```

### Removing a torrent
`DELETE /api/removetorrent`

```
{
    "infohash": "INFOHASH",
    "removefiles": false
}
```

### Streaming a file from torrent
`GET /api/stream`

```
/api/stream/:infohash/:filename
```

### Getting torrent stats
`GET /api/torrents`

For all torrents
```
/api/torrents
```

For specific torrent
```
/api/torrents/:infohash
```