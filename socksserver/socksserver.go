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

package socksserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"sshproxy/server"

	"github.com/armon/go-socks5"
	"golang.org/x/crypto/ssh"
)

type ResolverFunc func(ctx context.Context, name string) (context.Context, net.IP, error)

func (f ResolverFunc) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	return f(ctx, name)
}

const (
	SOCKSSERVER_DEBUG                  = false
	SOCKSSERVER_SSH_KEEPALIVE_TIMEOUT  = 7
	SOCKSSERVER_SSH_KEEPALIVE_INTERVAL = 30
	SOCKSSERVER_SSH_RECONNECT_INTERVAL = 5
)

/*
type DummyLock struct{}

func (d *DummyLock) Lock() {
}

func (d *DummyLock) Unlock() {
}

func (d *DummyLock) RLock() {
}

func (d *DummyLock) RUnlock() {
}
*/

type SocksServer struct {
	server          server.Server
	client          *ssh.Client
	keepAliveTicker *time.Ticker
	keepAliveCancel chan bool
	wg              sync.WaitGroup
	port            int
	listen          net.Listener
	lock            sync.RWMutex
}

func NewSocksServer(server server.Server, port int) *SocksServer {
	s := SocksServer{}
	s.server = server
	s.port = port
	return &s
}

func (s *SocksServer) IsConnected() (bool, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.client != nil, s.listen != nil
}

func (s *SocksServer) GetStatistic() (uint64, uint64, uint64, time.Time) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.client != nil {
		return s.server.GetStatistic()
	}
	return 0, 0, 0, time.Time{}
}

func (s *SocksServer) Stop() {
	bCon, _ := s.IsConnected()
	if !bCon {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.cancelKeepAlive()
	if s.listen != nil {
		s.listen.Close()
		s.listen = nil
	}
	s.client.Close()
}

func (s *SocksServer) Start() error {
	// SOCKS5 Server mit SSH Dialer
	dialer := func(ctx context.Context, network, addr string) (net.Conn, error) {
		fmt.Println("addr", addr)
		if s.client != nil {
			n, err := s.client.Dial(network, addr)
			if err != nil {
				s.doReconnect()
			}
			return n, err
		} else {
			return nil, errors.New("not connected")
		}
	}

	conf := &socks5.Config{
		Dial: dialer,
		Resolver: ResolverFunc(func(ctx context.Context, name string) (context.Context, net.IP, error) {
			return ctx, nil, nil
		}),
	}

	server, err := socks5.New(conf)
	if err != nil {
		return err
	}

	go func() {
		err := s.connect()
		if err != nil {
			s.trace("connect", "error", err)
		}
		s.lock.Lock()
		addr := fmt.Sprintf("0.0.0.0:%d", s.port)
		s.listen, err = net.Listen("tcp", addr)
		if err != nil {
			s.trace("listen", "error", err)
			s.lock.Unlock()
			return
		}
		s.lock.Unlock()
		err = server.Serve(s.listen)
		s.trace("listen and serve", "tcp", addr)
		if err != nil {
			s.trace("server", "error", err)
		}
	}()

	return nil
}

func (s *SocksServer) connect() error {
	bCon, _ := s.IsConnected()
	if bCon {
		return errors.New("already connected")
	}
	client, err := s.server.Connect()
	s.lock.Lock()
	defer s.lock.Unlock()

	s.client = client
	s.startKeepAlive()
	return err
}

func (s *SocksServer) disconnect() {
	bCon, _ := s.IsConnected()
	if !bCon {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stopKeepAlive()
	s.server.Disonnect(&s.client)
}

func (s *SocksServer) doReconnect() {
	s.trace("doReconnect")
	s.lock.Lock()
	defer s.lock.Unlock()
	s.stopKeepAlive()
	err := s.server.Reconnect(&s.client)
	t := SOCKSSERVER_SSH_KEEPALIVE_INTERVAL
	if err != nil {
		t = SOCKSSERVER_SSH_RECONNECT_INTERVAL
	}
	s.keepAliveTicker.Reset(time.Duration(t) * time.Second)
}

func (s *SocksServer) cancelKeepAlive() {
	if s.keepAliveTicker != nil {
		s.keepAliveTicker.Stop()
		s.keepAliveCancel <- true
		s.keepAliveTicker = nil
		s.wg.Wait()
	}
}

func (s *SocksServer) stopKeepAlive() {
	if s.keepAliveTicker != nil {
		s.keepAliveTicker.Stop()
	}
}

func (s *SocksServer) startKeepAlive() {
	s.stopKeepAlive()
	s.keepAliveTicker = time.NewTicker(time.Duration(SOCKSSERVER_SSH_KEEPALIVE_INTERVAL) * time.Second)
	s.keepAliveCancel = make(chan bool, 1)
	s.wg.Add(1)

	go s.doKeepAlive(s.keepAliveTicker.C, s.keepAliveCancel)
}

func (s *SocksServer) doKeepAlive(trigger <-chan time.Time, cancel <-chan bool) {
	for {
		select {
		case <-trigger:
			s.sendKeepAlive()
		case <-cancel:
			s.wg.Done()
			return
		}
	}
}

func (s *SocksServer) sendKeepAlive() {
	bCon, _ := s.IsConnected()
	if !bCon {
		s.doReconnect()
		return
	}

	donech := make(chan error, 1)
	go func() {
		bCon, _ := s.IsConnected()
		if bCon {
			_, _, err := s.client.SendRequest("keepalive@openssh.com", true, nil)
			donech <- err
		}
	}()

	var err error
	select {
	case err = <-donech:
		if err != nil {
			s.trace("keep alive error", "error", err)
		}
	case <-time.After(SOCKSSERVER_SSH_KEEPALIVE_TIMEOUT * time.Second):
		err = errors.New("keep live timeout")
		s.trace("keep alive timeout", "error", err)
	}
	if err != nil {
		s.doReconnect()
	}
}

func (s *SocksServer) trace(msg string, args ...any) {
	if SOCKSSERVER_DEBUG {
		slog.Info(msg, args...)
	}
}
