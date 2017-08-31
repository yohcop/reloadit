## Build

Clone this repo, then:

```
go get ./...
go build reload.go
```

## Use

From the directory you want to monitor and hot reload, just run the binary you
build above:

```
path/to/reload
```

This opens a webserver serving the static files in the directory on port 3000.
