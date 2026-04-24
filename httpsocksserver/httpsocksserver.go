// Copyright (c) 2026 Reiner Pröls
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//
// SPDX-License-Identifier: MIT
//
// Author: Reiner Pröls

package httpsocksserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/elazarl/goproxy"
	"golang.org/x/net/proxy"
)

const (
	HTTPSOCKSSERVER_DEBUG = false
)

type HttpSocksServer struct {
	server   *http.Server
	proxy    *goproxy.ProxyHttpServer
	httpPort int
}

func NewHttpSocksServer(socksAddr string, socksPort, httpPort int) (*HttpSocksServer, error) {
	h := HttpSocksServer{
		httpPort: httpPort,
	}
	dialer, err := proxy.SOCKS5("tcp", fmt.Sprintf("%s:%d", socksAddr, socksPort), nil, proxy.Direct)
	if err != nil {
		h.trace("socks5", "error", err)
		return nil, err
	}
	h.proxy = goproxy.NewProxyHttpServer()
	h.proxy.Verbose = HTTPSOCKSSERVER_DEBUG
	h.proxy.ConnectDial = dialer.Dial

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		},
		MaxIdleConns:        100,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		MaxIdleConnsPerHost: 20,
		TLSHandshakeTimeout: 10 * time.Second,
		ForceAttemptHTTP2:   true,
	}
	h.proxy.Tr = transport

	h.proxy.OnRequest().DoFunc(
		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			h.trace("REQUEST", "methode", req.Method, "url", req.URL)
			return req, nil
		},
	)
	return &h, nil
}

func (h *HttpSocksServer) Start() {
	h.server = &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", h.httpPort),
		Handler: h.proxy,
	}
	go func() {
		err := h.server.ListenAndServe()
		// https ListenAndServeTLS
		if err != nil {
			h.trace("Start", "error", err)
		}
	}()
}

func (h *HttpSocksServer) Stop() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := h.server.Shutdown(ctx)
		if err != nil {
			h.trace("Stop", "error", err)
		}
		h.server = nil
	}
}

func (s *HttpSocksServer) trace(msg string, args ...any) {
	if HTTPSOCKSSERVER_DEBUG {
		slog.Info(msg, args...)
	}
}
