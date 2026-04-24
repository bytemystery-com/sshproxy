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

package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

// Parameter
const (
	DEF_SALT_SIZE      = 16 // 32
	DEF_NONCE_SIZE     = 12 // (16?)
	DEF_KEY_SIZE       = 32
	DEF_ARGON2_TIME    = 5
	DEF_ARGON2_MEMORY  = 128
	DEF_ARGON2_THREADS = 4
)

type CryptCfg struct {
	SaltSize      int
	NonceSize     int
	KeySize       uint32
	Argon2Time    uint32
	Argon2Memory  uint32
	Argon2Threads uint8
}

type Crypt struct {
	Cfg CryptCfg
}

func NewCrypt(cfg *CryptCfg) *Crypt {
	c := Crypt{}
	if cfg != nil {
		c.Cfg = *cfg
	} else {
		c.Cfg = CryptCfg{
			SaltSize:      DEF_SALT_SIZE,
			KeySize:       DEF_KEY_SIZE,
			NonceSize:     DEF_NONCE_SIZE,
			Argon2Time:    DEF_ARGON2_TIME,
			Argon2Memory:  DEF_ARGON2_MEMORY,
			Argon2Threads: DEF_ARGON2_THREADS,
		}
	}
	return &c
}

// EncryptPassword verschlüsselt ein Passwort mit einem Masterpasswort
func (cry *Crypt) EncryptPassword(masterPassword, password string) (string, error) {
	b, err := cry.Encrypt([]byte(masterPassword), []byte(password))
	if err != nil {
		return "", err
	}
	// Base64 kodieren
	encoded := base64.StdEncoding.EncodeToString(b)
	return encoded, nil
}

func (cry *Crypt) Encrypt(masterPassword, data []byte) ([]byte, error) {
	// Salt generieren
	salt := make([]byte, cry.Cfg.SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Key ableiten
	key := argon2.IDKey(masterPassword, salt, cry.Cfg.Argon2Time, cry.Cfg.Argon2Memory*1024,
		cry.Cfg.Argon2Threads, cry.Cfg.KeySize)

	// AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Nonce generieren
	nonce := make([]byte, cry.Cfg.NonceSize)
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	// Verschlüsseln
	ciphertext := gcm.Seal(nil, nonce, []byte(data), nil)

	// Alles zusammenfügen: Salt + Nonce + Ciphertext
	dataOut := append(salt, nonce...)
	dataOut = append(dataOut, ciphertext...)

	return dataOut, nil
}

// DecryptPassword entschlüsselt den Base64-String wieder
func (cry *Crypt) DecryptPassword(masterPassword, encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	b, err := cry.Decrypt([]byte(masterPassword), data)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (cry *Crypt) Decrypt(masterPassword, encoded []byte) ([]byte, error) {
	if len(encoded) == 0 {
		return nil, errors.New("length 0")
	}

	// Salt, Nonce, Ciphertext extrahieren
	if (len(encoded)) < cry.Cfg.SaltSize+cry.Cfg.NonceSize {
		return nil, errors.New("data length too short")
	}
	salt := encoded[:cry.Cfg.SaltSize]
	nonce := encoded[cry.Cfg.SaltSize : cry.Cfg.SaltSize+cry.Cfg.NonceSize]
	ciphertext := encoded[cry.Cfg.SaltSize+cry.Cfg.NonceSize:]

	// Key ableiten
	key := argon2.IDKey(masterPassword, salt, cry.Cfg.Argon2Time, cry.Cfg.Argon2Memory*1024,
		cry.Cfg.Argon2Threads, cry.Cfg.KeySize)

	// AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Entschlüsseln
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

func (cry *Crypt) CreateKey(masterPassword []byte) ([]byte, error) {
	key := make([]byte, cry.Cfg.KeySize)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, err
	}
	d, err := cry.Encrypt([]byte(masterPassword), key)
	if err != nil {
		return nil, err
	}
	return d, err
}

func ErasePassword(pass []byte) {
	for i := range pass {
		pass[i] = 0
	}
}
