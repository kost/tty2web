// httpsc in Go. Copyright (C) Kost. Distributed under MIT.

package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// CmdReq holds JSON input request.
type CmdReq struct {
	Cmd    string
	Type   string
	Nojson bool
	Stdin  string
}

// CmdResp holds JSON output request.
type CmdResp struct {
	Cmd string
	Err string
}

type SCConfig struct {
	VerboseLevel int
	SilentOutput bool
	Log          *log.Logger
}

// real content Handler
func (cc *SCConfig) SCHandler(w http.ResponseWriter, r *http.Request) {
	var jsonout bool
	var inputjson CmdReq
	var outputjson CmdResp
	var body []byte
	ptype := "sc"

	if r.Header.Get("Content-Type") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		jsonout = true
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	if len(r.Header.Get("Accept-Language")) > 0 {
		ptype = r.Header.Get("Accept-Language")
	}
	cmdstr := ""
	urlq, urlErr := url.QueryUnescape(r.URL.RawQuery)
	if urlErr != nil {
		cc.Log.Printf("url query unescape: %v", urlErr)
	}
	if r.Method == "GET" || r.Method == "HEAD" {
		cmdstr = urlq
	}
	if r.Method == "POST" {
		var rerr error
		body, rerr = ioutil.ReadAll(r.Body)
		if rerr != nil {
			cc.Log.Printf("read Body: %v", rerr)
		}
		if closeErr := r.Body.Close(); closeErr != nil {
			cc.Log.Printf("body close: %v", closeErr)

		}
		if cc.VerboseLevel > 2 {
			cc.Log.Printf("Body: %s", body)
		}

		if len(urlq) > 0 {
			cmdstr = urlq
		} else {
			if jsonout {
				jerr := json.Unmarshal(body, &inputjson)
				if jerr != nil {
					// http.Error(w, jerr.Error(), 400)
					return
				}
				cmdstr = inputjson.Cmd
				ptype = inputjson.Type
				jsonout = !inputjson.Nojson
			} else {
				cmdstr = string(body)
			}
		}
	}
	if cc.VerboseLevel > 0 {
		log.Printf("Command to execute: %s", cmdstr)
	}

	if len(cmdstr) < 1 {
		return
	}

	if cc.VerboseLevel > 0 {
		log.Printf("Type to execute: %s", ptype)
	}

	if jsonout {
		outputjson.Cmd = cmdstr
		if encodeErr := json.NewEncoder(w).Encode(outputjson); encodeErr != nil {
			cc.Log.Printf("encode: %v", encodeErr)
		}
	}

	go ForkAndExecute(cc, ptype, cmdstr)
}
