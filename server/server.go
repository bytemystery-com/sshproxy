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

package server

import (
	"bytes"
	"errors"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

type countingConn struct {
	net.Conn
	readBytes  atomic.Uint64
	writeBytes atomic.Uint64
	reConnects atomic.Uint64
	startTime  time.Time
}

func (c *countingConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.readBytes.Add(uint64(n))
	return n, err
}

func (c *countingConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.writeBytes.Add(uint64(n))
	return n, err
}

func (c *countingConn) Reconnect() {
	c.reConnects.Add(1)
	c.startTime = time.Now()
}

type Server struct {
	Name             string                       `json:"name"`
	Host             string                       `json:"host"`
	Port             int                          `json:"port"`
	User             string                       `json:"user"`
	Password         string                       `json:"pass"`
	KeyFile          string                       `json:"keyfile"`
	KeyFileReader    func(string) ([]byte, error) `json:"-"`
	KeyFileContent   []byte                       `json:"keyfilecontent"`
	HostFiles        []string                     `json:"hostfiles"`
	HostFilesContent [][]byte                     `json:"hostfilescontent"`
	HostFileReader   func(string) ([]byte, error) `json:"-"`
	countConnection  *countingConn                `json:"-"`
}

func (server *Server) IsAlive() bool {
	conn, err := net.Dial("tcp", net.JoinHostPort(server.Host, strconv.Itoa(server.Port)))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func (server *Server) Reconnect(client **ssh.Client) error {
	if client == nil {
		return errors.New("client is nil")
	}
	if *client != nil {
		server.Disonnect(client)
	}
	c, err := server.Connect()
	if err != nil {
		return err
	}
	*client = c
	server.countConnection.Reconnect()
	return nil
}

func (server *Server) Disonnect(client **ssh.Client) error {
	if client == nil {
		return errors.New("client is nil")
	}
	if *client == nil {
		return errors.New("client is nil")
	}
	err := (*client).Close()
	*client = nil
	return err
}

func (server *Server) Connect() (*ssh.Client, error) {
	var client *ssh.Client = nil
	// var err error
	if len(server.Host) > 0 {
		var sig ssh.Signer = nil
		if len(server.KeyFile) > 0 || len(server.KeyFileContent) > 0 {
			var key []byte
			var err error
			if len(server.KeyFileContent) > 0 {
				key = server.KeyFileContent
			} else {
				if server.KeyFileReader != nil {
					key, err = server.KeyFileReader(server.KeyFile)
				} else {
					key, err = os.ReadFile(server.KeyFile)
				}
			}
			if err != nil {
				return nil, err
			}
			if len(server.Password) > 0 {
				sig, err = ssh.ParsePrivateKeyWithPassphrase([]byte(key), []byte(server.Password))
			} else {
				sig, err = ssh.ParsePrivateKey([]byte(key))
			}

			if err != nil {
				return nil, err
			}
		}
		var hostKeys []ssh.PublicKey
		if len(server.HostFiles) > 0 || len(server.HostFilesContent) > 0 {
			if len(server.HostFilesContent) > 0 {
				for _, host := range server.HostFilesContent {
					hKey, _, _, _, err := ssh.ParseAuthorizedKey(host)
					if err == nil {
						hostKeys = append(hostKeys, hKey)
					}
				}
			} else {
				hostKeys = make([]ssh.PublicKey, 0, len(server.HostFiles))
				for _, item := range server.HostFiles {
					var host []byte
					var err error
					if server.HostFileReader != nil {
						host, err = server.HostFileReader(item)
					} else {
						host, err = os.ReadFile(item)
					}
					if err == nil {
						hKey, _, _, _, err := ssh.ParseAuthorizedKey(host)
						if err == nil {
							hostKeys = append(hostKeys, hKey)
						}
					}
				}
			}
		}

		auth := []ssh.AuthMethod{}
		if sig != nil {
			auth = append(auth, ssh.PublicKeys(sig))
		} else if len(server.KeyFile) == 0 {
			auth = append(auth, ssh.Password(server.Password), ssh.KeyboardInteractive(
				func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
					answers = make([]string, len(questions))
					for n := range questions {
						answers[n] = server.Password
					}
					return answers, nil
				}))
		}

		config := &ssh.ClientConfig{
			User: server.User,
			Auth: auth,
			/*
				HostKeyAlgorithms: []string{
					ssh.KeyAlgoED25519,
				},
			*/
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				if len(hostKeys) > 0 {
					for _, item := range hostKeys {
						if bytes.Equal(key.Marshal(), item.Marshal()) {
							return nil
						}
					}
					return errors.New("hostkey does not match")
				} else {
					return nil
				}
			},
		}
		rawConn, err := net.Dial("tcp", server.Host+":"+strconv.Itoa(server.Port))
		if err != nil {
			return nil, err
		}
		if server.countConnection == nil {
			server.countConnection = &countingConn{
				Conn: rawConn,
			}
		} else {
			server.countConnection.Conn = rawConn
		}
		sshConn, chans, reqs, err := ssh.NewClientConn(server.countConnection, server.Host+":"+strconv.Itoa(server.Port), config)
		if err != nil {
			return nil, err
		}
		client = ssh.NewClient(sshConn, chans, reqs)

		if err != nil {
			return nil, err
		}
	} else {
		client = nil
	}
	server.countConnection.startTime = time.Now()
	return client, nil
}

func (server *Server) GetStatistic() (uint64, uint64, uint64, time.Time) {
	if server.countConnection != nil {
		return server.countConnection.readBytes.Load(), server.countConnection.writeBytes.Load(), server.countConnection.reConnects.Load(), server.countConnection.startTime
	} else {
		return 0, 0, 0, time.Time{}
	}
}
