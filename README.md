# go-openiris-esdif-cameras
Simple tool to use OpenIris esdif cameras over http on linux.

# Usage
Code gonna output everything to `localhost:8000` (or you can specify it using `-port`), to get all streams that you can use, you can just go directly to this URL in your browser.

Stream URL Made from device address that obtained from your system.

To use it in Baballonia on linux you need to use `IpCameraCapture` backend.

# How to build
Just clone repository, install golang >= 1.24.5 and run `go build` inside directory.
It gonna produce `go-openiris-esdif-cameras` executable that you can simply run and it gonna stream it on default url.

# Special thanks
- [go-bsb-cams](https://github.com/LilliaElaine/go-bsb-cams/tree/main) for workaround that works in same way.
