package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"regexp"
	noesctmpl "text/template"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/kost/tty2web/pkg/homedir"
	"github.com/kost/tty2web/pkg/randomstring"
	"github.com/kost/tty2web/webtty"
	"fmt"
	"bufio"
	"bytes"
	"encoding/base64"
	"github.com/hashicorp/yamux"
	"strings"
	ntlmssp "github.com/kost/go-ntlmssp"
)

var encBase64 = base64.StdEncoding.EncodeToString
var decBase64 = base64.StdEncoding.DecodeString
var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response
var session *yamux.Session

func connectviaproxy(proxyaddr string, connectaddr string, proxyauth string) (net.Conn, error) {
	var username string
	var domain string
	var password string
	var useragent string
	connectproxystring := ""
	socksdebug:=true
	proxyauthstring := &proxyauth
	var dummyConn net.Conn

	if strings.Contains(*proxyauthstring, "/") {
		domain = strings.Split(*proxyauthstring, "/")[0]
		username = strings.Split(strings.Split(*proxyauthstring, "/")[1], ":")[0]
		password = strings.Split(strings.Split(*proxyauthstring, "/")[1], ":")[1]
	} else {
		username = strings.Split(*proxyauthstring, ":")[0]
		password = strings.Split(*proxyauthstring, ":")[1]
	}
	log.Printf("Using domain %s with %s:%s", domain, username, password)

	if (username != "") && (password != "") && (domain != "") {
		negotiateMessage, errn := ntlmssp.NewNegotiateMessage(domain, "")
		if errn != nil {
			log.Println("NEG error")
			log.Println(errn)
			return dummyConn, errn
		}
		log.Print(negotiateMessage)
		negheader := fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiateMessage))

		connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
			"\r\nUser-Agent: " + useragent +
			"\r\nProxy-Authorization: " + negheader +
			"\r\nProxy-Connection: Keep-Alive" +
			"\r\n\r\n"

	} else {
		connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
			"\r\nUser-Agent: " + useragent +
			"\r\nProxy-Connection: Keep-Alive" +
			"\r\n\r\n"
	}

	if socksdebug {
		log.Print(connectproxystring)
	}

	conn, err := net.Dial("tcp", proxyaddr)
	if err != nil {
		// handle error
		log.Printf("Error connect: %v", err)
		return dummyConn, err
	}
	conn.Write([]byte(connectproxystring))

	time.Sleep(proxytimeout) //Because socket does not close - we need to sleep for full response from proxy

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	status := resp.Status

	if socksdebug {
		log.Print(status)
		log.Print(resp)
	}

	if (resp.StatusCode == 200) || (strings.Contains(status, "HTTP/1.1 200 ")) ||
		(strings.Contains(status, "HTTP/1.0 200 ")) {
		log.Print("Connected via proxy. No auth required")
		return conn, nil
	}

	if socksdebug {
		log.Print("Checking proxy auth")
	}
	if resp.StatusCode == 407 {
		log.Print("Got Proxy status code (407)")
		ntlmchall := resp.Header.Get("Proxy-Authenticate")
		log.Print(ntlmchall)
		if strings.Contains(ntlmchall, "NTLM") {
			if socksdebug {
				log.Print("Got NTLM challenge:")
				log.Print(ntlmchall)
			}

			/*
				negstring:= fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiateMessage))
				connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
					"\r\nUser-Agent: "+useragent+
					"\r\nProxy-Authorization: " + negstring +
					"\r\nProxy-Connection: Keep-Alive" +
					"\r\n\r\n"
			*/

			ntlmchall = ntlmchall[5:]
			if socksdebug {
				log.Print("NTLM challenge:")
				log.Print(ntlmchall)
			}
			challengeMessage, errb := decBase64(ntlmchall)
			if errb != nil {
				log.Println("BASE64 Decode error")
				log.Println(errb)
				return dummyConn, errb
			}
			authenticateMessage, erra := ntlmssp.ProcessChallenge(challengeMessage, username, password)
			if erra != nil {
				log.Println("Process challenge error")
				log.Println(erra)
				return dummyConn, erra
			}

			authMessage := fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticateMessage))

			//log.Print(authenticate)
			connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
				"\r\nUser-Agent: Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko" +
				"\r\nProxy-Authorization: " + authMessage +
				"\r\nProxy-Connection: Keep-Alive" +
				"\r\n\r\n"
		} else if strings.Contains(ntlmchall, "Basic") {
			if socksdebug {
				log.Print("Got Basic challenge:")
			}
			var authbuffer bytes.Buffer
			authbuffer.WriteString(username)
			authbuffer.WriteString(":")
			authbuffer.WriteString(password)

			basicauth := encBase64(authbuffer.Bytes())

			//log.Print(authenticate)
			connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
				"\r\nUser-Agent: Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko" +
				"\r\nProxy-Authorization: Basic " + basicauth +
				"\r\nProxy-Connection: Keep-Alive" +
				"\r\n\r\n"
		} else {
			log.Print("Unknown authentication")
			return dummyConn, errors.New("Unknown authentication")
		}
		log.Print("Connecting to proxy")
		log.Print(connectproxystring)
		conn.Write([]byte(connectproxystring))

		//read response
		bufReader := bufio.NewReader(conn)
		conn.SetReadDeadline(time.Now().Add(proxytimeout))
		statusb, _ := ioutil.ReadAll(bufReader)

		status = string(statusb)

		//disable socket read timeouts
		conn.SetReadDeadline(time.Now().Add(100 * time.Hour))

		if strings.Contains(status, "HTTP/1.1 200 ") {
			log.Print("Connected via proxy")
			return conn, nil
		}
		log.Printf("Not Connected via proxy. Status:%v", status)
		return dummyConn, errors.New("Not connected via proxy")

	}
	log.Print("Not connected via proxy")
	conn.Close()
	return dummyConn, nil
}

func connectForSocks(address string, proxy string, proxyauth string, agentpassword string) (*yamux.Session, error) {
	var err error
	var yam *yamux.Session

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	var conn net.Conn
	var connp net.Conn
	var newconn net.Conn
	//var conntls tls.Conn
	//var conn tls.Conn
	if proxy == "" {
		log.Println("Connecting to far end")
		//conn, err = net.Dial("tcp", address)
		conn, err = tls.Dial("tcp", address, conf)
		if err != nil {
			log.Printf("Cannot connect to %s: %s", address, err)
			return yam, err
		}
	} else {
		log.Println("Connecting to proxy ...")
		connp, err = connectviaproxy(proxy, address, proxyauth)
		if err != nil {
			log.Println("Proxy successfull. Connecting to far end")
			conntls := tls.Client(connp, conf)
			err := conntls.Handshake()
			if err != nil {
				log.Printf("Error connect: %v", err)
				return yam,err
			}
			newconn = net.Conn(conntls)
		} else {
			log.Println("Proxy NOT successfull. Exiting")
			return yam, err
		}
	}

	log.Println("Starting client")
	if proxy == "" {
		conn.Write([]byte(agentpassword))
		//time.Sleep(time.Second * 1)
		session, err = yamux.Server(conn, nil)
	} else {

		//log.Print(conntls)
		newconn.Write([]byte(agentpassword))
		time.Sleep(time.Second * 1)
		session, err = yamux.Server(newconn, nil)
	}
	if err != nil {
		log.Println("Error session")
		return yam, err
	}
	log.Println("Returning session")
	return session, err
}

// Server provides a webtty HTTP endpoint.
type Server struct {
	factory Factory
	options *Options

	upgrader      *websocket.Upgrader
	indexTemplate *template.Template
	titleTemplate *noesctmpl.Template
}

// New creates a new instance of Server.
// Server will use the New() of the factory provided to handle each request.
func New(factory Factory, options *Options) (*Server, error) {
	indexData, err := Asset("static/index.html")
	if err != nil {
		panic("index not found") // must be in bindata
	}
	if options.IndexFile != "" {
		path := homedir.Expand(options.IndexFile)
		indexData, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read custom index file at `%s`", path)
		}
	}
	indexTemplate, err := template.New("index").Parse(string(indexData))
	if err != nil {
		panic("index template parse failed") // must be valid
	}

	titleTemplate, err := noesctmpl.New("title").Parse(options.TitleFormat)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse window title format `%s`", options.TitleFormat)
	}

	var originChekcer func(r *http.Request) bool
	if options.WSOrigin != "" {
		matcher, err := regexp.Compile(options.WSOrigin)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to compile regular expression of Websocket Origin: %s", options.WSOrigin)
		}
		originChekcer = func(r *http.Request) bool {
			return matcher.MatchString(r.Header.Get("Origin"))
		}
	}

	return &Server{
		factory: factory,
		options: options,

		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Subprotocols:    webtty.Protocols,
			CheckOrigin:     originChekcer,
		},
		indexTemplate: indexTemplate,
		titleTemplate: titleTemplate,
	}, nil
}

// Run starts the main process of the Server.
// The cancelation of ctx will shutdown the server immediately with aborting
// existing connections. Use WithGracefullContext() to support gracefull shutdown.
func (server *Server) Run(ctx context.Context, options ...RunOption) error {
	cctx, cancel := context.WithCancel(ctx)
	opts := &RunOptions{gracefullCtx: context.Background()}
	for _, opt := range options {
		opt(opts)
	}

	counter := newCounter(time.Duration(server.options.Timeout) * time.Second)

	path := "/"
	if server.options.EnableRandomUrl {
		path = "/" + randomstring.Generate(server.options.RandomUrlLength) + "/"
	}

	handlers := server.setupHandlers(cctx, cancel, path, counter)
	srv, err := server.setupHTTPServer(handlers)
	if err != nil {
		return errors.Wrapf(err, "failed to setup an HTTP server")
	}

	if server.options.PermitWrite {
		log.Printf("Permitting clients to write input to the PTY.")
	}
	if server.options.Once {
		log.Printf("Once option is provided, accepting only one client")
	}

	if server.options.Port == "0" {
		log.Printf("Port number configured to `0`, choosing a random port")
	}

	srvErr := make(chan error, 1)

	if server.options.Connect == "" {
		hostPort := net.JoinHostPort(server.options.Address, server.options.Port)
		listener, err := net.Listen("tcp", hostPort)
		if err != nil {
			return errors.Wrapf(err, "failed to listen at `%s`", hostPort)
		}

		scheme := "http"
		if server.options.EnableTLS {
			scheme = "https"
		}
		host, port, _ := net.SplitHostPort(listener.Addr().String())
		log.Printf("HTTP server is listening at: %s", scheme+"://"+host+":"+port+path)
		if server.options.Address == "0.0.0.0" {
			for _, address := range listAddresses() {
				log.Printf("Alternative URL: %s", scheme+"://"+address+":"+port+path)
			}
		}
		go func() {
			if server.options.EnableTLS {
				crtFile := homedir.Expand(server.options.TLSCrtFile)
				keyFile := homedir.Expand(server.options.TLSKeyFile)
				log.Printf("TLS crt file: " + crtFile)
				log.Printf("TLS key file: " + keyFile)

				err = srv.ServeTLS(listener, crtFile, keyFile)
			} else {
				err = srv.Serve(listener)
			}
			if err != nil {
				srvErr <- err
			}
		}()
	} else {
		go func() {
			session, err = connectForSocks(server.options.Connect,server.options.Proxy, server.options.ProxyAuth, server.options.Password)
			if err != nil {
				log.Printf("Error creating sessions %s", err)
				srvErr <- err
				return
			}
			err = srv.Serve(session)
			if err != nil {
				srvErr <- err
			}
		}()
	}

	go func() {
		select {
		case <-opts.gracefullCtx.Done():
			srv.Shutdown(context.Background())
		case <-cctx.Done():
		}
	}()

	select {
	case err = <-srvErr:
		if err == http.ErrServerClosed { // by gracefull ctx
			err = nil
		} else {
			cancel()
		}
	case <-cctx.Done():
		srv.Close()
		err = cctx.Err()
	}

	conn := counter.count()
	if conn > 0 {
		log.Printf("Waiting for %d connections to be closed", conn)
	}
	counter.wait()

	return err
}

func (server *Server) setupHandlers(ctx context.Context, cancel context.CancelFunc, pathPrefix string, counter *counter) http.Handler {
	staticFileHandler := http.FileServer(
		&assetfs.AssetFS{Asset: Asset, AssetDir: AssetDir, Prefix: "static"},
	)

	var siteMux = http.NewServeMux()
	siteMux.HandleFunc(pathPrefix, server.handleIndex)
	siteMux.Handle(pathPrefix+"js/", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"favicon.png", http.StripPrefix(pathPrefix, staticFileHandler))
	siteMux.Handle(pathPrefix+"css/", http.StripPrefix(pathPrefix, staticFileHandler))

	siteMux.HandleFunc(pathPrefix+"auth_token.js", server.handleAuthToken)
	siteMux.HandleFunc(pathPrefix+"config.js", server.handleConfig)

	siteHandler := http.Handler(siteMux)

	if server.options.EnableBasicAuth {
		log.Printf("Using Basic Authentication")
		siteHandler = server.wrapBasicAuth(siteHandler, server.options.Credential)
	}

	withGz := gziphandler.GzipHandler(server.wrapHeaders(siteHandler))
	siteHandler = server.wrapLogger(withGz)

	wsMux := http.NewServeMux()
	wsMux.Handle("/", siteHandler)
	wsMux.HandleFunc(pathPrefix+"ws", server.generateHandleWS(ctx, cancel, counter))
	siteHandler = http.Handler(wsMux)

	return siteHandler
}

func (server *Server) setupHTTPServer(handler http.Handler) (*http.Server, error) {
	srv := &http.Server{
		Handler: handler,
	}

	if server.options.EnableTLSClientAuth {
		tlsConfig, err := server.tlsConfig()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to setup TLS configuration")
		}
		srv.TLSConfig = tlsConfig
	}

	return srv, nil
}

func (server *Server) tlsConfig() (*tls.Config, error) {
	caFile := homedir.Expand(server.options.TLSCACrtFile)
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, errors.New("could not open CA crt file " + caFile)
	}
	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("could not parse CA crt file data in " + caFile)
	}
	tlsConfig := &tls.Config{
		ClientCAs:  caCertPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}
	return tlsConfig, nil
}
