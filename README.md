# torrenttp
Simple REST API controlled BitTorrent client

## Docker (deprecated)
```
docker run -d \
--name torrenttp \
-p 1010:1010 \ # Change the 1010 on the left to change the listening port
-v ~/tttpdl:/dl \ # Change the ~/tttpdl to the download path
-e NOUP=false \ # Set to true to disable uploads
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

### Uploading a torrent file
`POST /api/addtorrentfile`

Attach the file in the `torrent` field in `multipart/form-data`

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

### Download a file from torrent
`GET /api/file`

```
/api/file/:infohash/:filename
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