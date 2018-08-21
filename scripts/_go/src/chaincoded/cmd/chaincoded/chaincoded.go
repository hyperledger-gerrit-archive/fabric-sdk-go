/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	defaultDockerHost = "http://localhost:2375"
)

var (
	binaryRegExp         = regexp.MustCompile("(.*)_(.*)")
	containerNameRegEx   = regexp.MustCompile("(.*)-(.*)-(.*)-(.*)")
	containerStartRegEx  = regexp.MustCompile("/containers/(.+)/start")
	containerUploadRegEx = regexp.MustCompile("/containers/(.+)/archive")
)

var peerEndpoints map[string]string

type chaincoded struct {
	proxy *httputil.ReverseProxy
	wg    *sync.WaitGroup
	done  chan struct{}
}

type chaincodeParams struct {
	network   string
	hostname  string
	ccID      string
	ccVersion string
}

func newChaincoded(wg *sync.WaitGroup, done chan struct{}) *chaincoded {
	docker_host, ok := os.LookupEnv("DOCKER_HOST")
	if !ok {
		docker_host = defaultDockerHost
	}

	url, err := url.Parse(docker_host)
	if err != nil {
		logFatalf("invalid URL for docker host: %s", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	d := chaincoded{
		proxy: proxy,
		wg:    wg,
		done:  done,
	}

	return &d
}

func launchChaincode(ccParams *chaincodeParams, tlsPath string, done chan struct{}) {
	const (
		relaunchWaitTime = time.Second
	)

	for {
		logDebugf("Starting chaincode [%s, %s]", ccParams.chaincodeBinary(), ccParams.ccID)
		cmd := createChaincodeCmd(ccParams, tlsPath)

		cmdDone := make(chan struct{})

		go func() {
			defer close(cmdDone)

			err := cmd.Run()
			if v, ok := err.(*exec.ExitError); ok {
				logWarningf("Chaincode had an exit error - will relaunch [%s, %s]", ccParams.ccID, v)
			} else {
				logFatalf("Chaincode had a non-exit error [%s, %s]", ccParams.ccID, err)
			}
		}()

		select {
		case <-done:
			logDebugf("Stopping chaincode [%s, %s, %d]", ccParams.chaincodeBinary(), ccParams.ccID, cmd.Process.Pid)
			err := cmd.Process.Kill()
			if err != nil {
				logWarningf("Killing process failed [%s, %s]", ccParams.ccID, err)
			}
			return
		case <-cmdDone:
			time.Sleep(relaunchWaitTime)
		}
	}
}

func createChaincodeCmd(ccParams *chaincodeParams, tlsPath string) *exec.Cmd {
	rootCertFile := path.Join(tlsPath, "peer.crt")
	keyPath := path.Join(tlsPath, "client.key")
	certPath := path.Join(tlsPath, "client.crt")

	cmd := exec.Command(ccParams.chaincodeBinary(), tlsPath)
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"CORE_PEER_ADDRESS="+ccParams.peerAddr(),
		"CORE_CHAINCODE_ID_NAME="+ccParams.chaincodeID(),
		"CORE_PEER_TLS_ENABLED=TRUE",
		"CORE_PEER_TLS_ROOTCERT_FILE="+rootCertFile,
		"CORE_TLS_CLIENT_KEY_PATH="+keyPath,
		"CORE_TLS_CLIENT_CERT_PATH="+certPath,
		"CORE_CHAINCODE_LOGGING_LEVEL="+getChaincodeLoggingLevel(),
	)

	// Chaincode and shim logs are output through Stderr.
	// Some chaincodes also print messages to Stdout.
	if isVerbose() {
		cmd.Stdout = os.Stdout
	}

	return cmd
}

func extractChaincodeParams(containerName string) (*chaincodeParams, bool) {
	m := containerNameRegEx.FindStringSubmatch(containerName)

	if m == nil {
		return nil, false
	}

	ccParams := chaincodeParams{
		network:   m[1],
		hostname:  m[2],
		ccID:      m[3],
		ccVersion: m[4],
	}

	return &ccParams, true
}

func (cp *chaincodeParams) chaincodeID() string {
	return cp.ccID + ":" + cp.ccVersion
}

func (cp *chaincodeParams) chaincodeBinary() string {
	m := binaryRegExp.FindStringSubmatch(cp.ccID)

	if m == nil {
		return "example_cc"
	}

	return m[1]
}

func (cp *chaincodeParams) peerAddr() string {
	endpoint, ok := peerEndpoints[cp.hostname]
	if !ok {
		return "localhost:7052"
	}

	return endpoint
}

func (d *chaincoded) handleUploadToContainerRequest(w http.ResponseWriter, r *http.Request, containerName string) {
	ccParams, ok := extractChaincodeParams(containerName)
	if !ok {
		d.proxy.ServeHTTP(w, r)
		return
	}

	logDebugf("Handling upload to container request [%s]", containerName)
	tmpDir, err := ioutil.TempDir("", "chaincoded")
	if err != nil {
		logWarningf("creation of temporary directory failed: %s", err)
		w.WriteHeader(500)
		return
	}

	err = extractArchive(r.Body, tmpDir)
	if err != nil {
		logWarningf("extracting archive failed: %s", err)
		w.WriteHeader(500)
		return
	}

	tlsPath := path.Join(tmpDir, "etc", "hyperledger", "fabric")
	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		launchChaincode(ccParams, tlsPath, d.done)
	}()

	w.WriteHeader(200)
}

func (d *chaincoded) handleStartContainerRequest(w http.ResponseWriter, r *http.Request, containerName string) {
	logDebugf("Handling start container request [%s]", containerName)
	w.WriteHeader(204)
}

func extractArchive(in io.Reader, basePath string) error {
	gr, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			outPath := path.Join(basePath, hdr.Name)

			if err := os.Mkdir(outPath, 0755); err != nil {
				return err
			}
		case 0:
			fallthrough
		case tar.TypeReg:
			outPath := path.Join(basePath, hdr.Name)

			dir := path.Dir(outPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}

			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func (d *chaincoded) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	startMatches := containerStartRegEx.FindStringSubmatch(r.URL.Path)
	uploadMatches := containerUploadRegEx.FindStringSubmatch(r.URL.Path)

	logDebugf("Handling HTTP request [%s]", r.URL)
	if startMatches != nil {
		d.handleStartContainerRequest(w, r, startMatches[1])
	} else if uploadMatches != nil {
		d.handleUploadToContainerRequest(w, r, uploadMatches[1])
	} else {
		d.proxy.ServeHTTP(w, r)
	}
}

func runHTTPServer(addr string, dh *chaincoded, wg *sync.WaitGroup, done chan struct{}) {
	srv := &http.Server{Addr: addr, Handler: dh}

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			logFatalf("HTTP server failed [%s]", err)
		}
		logDebugf("HTTP server stopped [%s]", addr)
	}()

	// wait for signal to exit and then gracefully shutdown
	<-done
	srv.Shutdown(context.Background())
}

func waitForTermination() {
	s := make(chan os.Signal, 1)
	signal.Notify(s,
		syscall.SIGINT,
		syscall.SIGTERM)

	<-s
}

func main() {

	const cmdHelp = "arguments are the listen addr followed by a list of chaincode endpoints (hostname:port)"
	peerEndpoints = make(map[string]string)

	if len(os.Args) < 2 {
		log.Fatal(cmdHelp)
	}

	addr := os.Args[1]
	logInfof("Chaincoded starting [%s] ...", addr)

	for _, endpoint := range os.Args[2:] {
		s := strings.Split(endpoint, ":")
		if len(s) != 2 {
			log.Fatal(cmdHelp)
		}
		peerEndpoints[s[0]] = endpoint
	}

	var wg sync.WaitGroup
	done := make(chan struct{})

	dh := newChaincoded(&wg, done)
	go runHTTPServer(addr, dh, &wg, done)

	waitForTermination()
	logInfof("Chaincoded stopping [%s] ...", addr)
	close(done)
	wg.Wait()
	logInfof("Chaincoded stopped [%s]", addr)
}
