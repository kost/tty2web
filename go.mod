module github.com/kost/tty2web

go 1.23.0

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/creack/pty v1.1.20
	github.com/fatih/structs v1.1.0
	github.com/gorilla/websocket v1.5.1
	github.com/hashicorp/yamux v0.1.1
	github.com/kost/dnstun v0.0.0-20230511164951-6e7f5656a900
	github.com/kost/go-ntlmssp v0.0.0-20190601005913-a22bdd33b2a4
	github.com/kost/gosc v0.0.0-20230110210303-490723ad1528
	github.com/kost/httpexecute v0.0.0-20211119174050-f41d120e9db6
	github.com/kost/regeorgo v0.0.0-20211119151427-d6c70e76b00e
	github.com/pkg/errors v0.9.1
	github.com/urfave/cli/v2 v2.25.7
	github.com/yudai/hcl v0.0.0-20151013225006-5fa2393b3552
	golang.org/x/crypto v0.35.0
)

require (
	github.com/Jeffail/tunny v0.1.4 // indirect
	github.com/acomagu/bufpipe v1.0.4 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kost/chashell v0.0.0-20230409212000-cf0fbd106275 // indirect
	github.com/miekg/dns v1.1.56 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.21.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/tools v0.14.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace (
	github.com/creack/pty => github.com/photostorm/pty v1.1.19-0.20221026012344-0a71ca4f0f8c
	github.com/creack/pty v1.1.18 => github.com/photostorm/pty v1.1.19-0.20221026012344-0a71ca4f0f8c
)
