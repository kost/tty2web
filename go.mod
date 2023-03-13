module github.com/kost/tty2web

go 1.18

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/creack/pty v1.1.18
	github.com/elazarl/go-bindata-assetfs v1.0.1
	github.com/fatih/structs v1.1.0
	github.com/gorilla/websocket v1.5.0
	github.com/hashicorp/yamux v0.1.1
	github.com/kost/go-ntlmssp v0.0.0-20190601005913-a22bdd33b2a4
	github.com/kost/gosc v0.0.0-20230110210303-490723ad1528
	github.com/kost/httpexecute v0.0.0-20211119174050-f41d120e9db6
	github.com/kost/regeorgo v0.0.0-20211119151427-d6c70e76b00e
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli/v2 v2.25.0
	github.com/yudai/hcl v0.0.0-20151013225006-5fa2393b3552
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

replace (
	github.com/creack/pty => github.com/photostorm/pty v1.1.19-0.20221026012344-0a71ca4f0f8c
	github.com/creack/pty v1.1.18 => github.com/photostorm/pty v1.1.19-0.20221026012344-0a71ca4f0f8c
)
