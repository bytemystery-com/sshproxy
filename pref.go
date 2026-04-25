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

package main

import (
	"encoding/json"

	"sshproxy/crypt"
)

const (
	PREF_THEMEVARIANT_KEY            = "theme"
	PREF_THEMEVARIANT_VALUE          = -1
	PREF_PROXYLIST_KEY               = "proxylist"
	PREF_PROXYLIST_VALUE             = ""
	PREF_UIUPDATETIME_KEY            = "uiupdatetime"
	PREF_UIUPDATETIME_VALUE          = 250
	PREF_MASTERKEY_TEST_KEY          = "mastertest"
	PREF_MASTERKEY_TEST_VALUE        = "Reiner"
	PREF_UPDATE_LAST_CHECK_KEY       = "lastupdatecheck"
	PREF_UPDATE_LAST_CHECK_VALUE     = 0
	PREF_UPDATE_CHECK_INTERVAL_KEY   = "updatecheckinterval"
	PREF_UPDATE_CHECK_INTERVAL_VALUE = 48
	PREF_UPDATE_CHECK_AUTO_KEY       = "autoupdatecheck"
	PREF_UPDATE_CHECK_AUTO_VALUE     = true
	PREF_TASKS_FIRST_START_KEY       = "firststart"
	PREF_LAST_SELPROXY_INDEX_KEY     = "latselproxyindex"
	PREF_LAST_SELPROXY_INDEX_VALUE   = 0
	InternPassword                   = "425t785ßGuo34398)(453ß98$§\"56989879)()(/)(/87fknrofgIUHHGUIGuo"
)

type Preferences struct {
	ThemeVariant        int
	ProxyList           string // json String
	Crypt               crypt.Crypt
	UIUpdateTime        int // msec
	MasterKeyTest       string
	FirstStart          bool
	LastUpdatecheck     int64
	UpdateCheckInterval int
	AutoUpdateCheck     bool
	LastSelProxyIndex   int
}

func NewPreferences() *Preferences {
	p := &Preferences{
		ThemeVariant:        Gui.App.Preferences().IntWithFallback(PREF_THEMEVARIANT_KEY, PREF_THEMEVARIANT_VALUE),
		ProxyList:           Gui.App.Preferences().StringWithFallback(PREF_PROXYLIST_KEY, PREF_PROXYLIST_VALUE),
		Crypt:               *crypt.NewCrypt(nil),
		UIUpdateTime:        Gui.App.Preferences().IntWithFallback(PREF_UIUPDATETIME_KEY, PREF_UIUPDATETIME_VALUE),
		MasterKeyTest:       Gui.App.Preferences().StringWithFallback(PREF_MASTERKEY_TEST_KEY, PREF_MASTERKEY_TEST_VALUE),
		FirstStart:          Gui.App.Preferences().BoolWithFallback(PREF_TASKS_FIRST_START_KEY, true),
		LastUpdatecheck:     100 * int64(Gui.App.Preferences().IntWithFallback(PREF_UPDATE_LAST_CHECK_KEY, PREF_UPDATE_LAST_CHECK_VALUE)),
		UpdateCheckInterval: Gui.App.Preferences().IntWithFallback(PREF_UPDATE_CHECK_INTERVAL_KEY, PREF_UPDATE_CHECK_INTERVAL_VALUE),
		AutoUpdateCheck:     Gui.App.Preferences().BoolWithFallback(PREF_UPDATE_CHECK_AUTO_KEY, PREF_UPDATE_CHECK_AUTO_VALUE),
		LastSelProxyIndex:   Gui.App.Preferences().IntWithFallback(PREF_LAST_SELPROXY_INDEX_KEY, PREF_LAST_SELPROXY_INDEX_VALUE),
	}
	return p
}

func (p *Preferences) Store() {
	pref := Gui.App.Preferences()
	pref.SetInt(PREF_THEMEVARIANT_KEY, p.ThemeVariant)
	pref.SetString(PREF_PROXYLIST_KEY, p.ProxyList)
	pref.SetInt(PREF_UIUPDATETIME_KEY, p.UIUpdateTime)
	pref.SetString(PREF_MASTERKEY_TEST_KEY, p.MasterKeyTest)
	pref.SetBool(PREF_TASKS_FIRST_START_KEY, p.FirstStart)
	pref.SetInt(PREF_UPDATE_LAST_CHECK_KEY, int(p.LastUpdatecheck/100))
	pref.SetInt(PREF_UPDATE_CHECK_INTERVAL_KEY, p.UpdateCheckInterval)
	pref.SetBool(PREF_UPDATE_CHECK_AUTO_KEY, p.AutoUpdateCheck)
	pref.SetInt(PREF_LAST_SELPROXY_INDEX_KEY, p.LastSelProxyIndex)
}

func (p *Preferences) SetProxies(plist []*ProxyEntry) error {
	pass, err := p.Crypt.DecryptPassword(InternPassword, Gui.MasterPassword)
	if err != nil {
		return err
	}
	list := make([]ProxyEntry, 0, len(plist))
	for _, item := range plist {
		entry := *item
		var err error
		entry.Password, err = p.Crypt.EncryptPassword(string(pass), entry.Password)
		if err != nil {
			return err
		}
		if len(entry.KeyFileContent) > 0 {
			entry.KeyFileContent, err = p.Crypt.Encrypt([]byte(pass), entry.KeyFileContent)
			if err != nil {
				return err
			}
		}
		list = append(list, entry)
	}
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	Gui.Settings.ProxyList = string(data)
	return nil
}

func (p *Preferences) GetProxyList() ([]*ProxyEntry, error) {
	pass, err := p.Crypt.DecryptPassword(InternPassword, Gui.MasterPassword)
	if err != nil {
		return nil, err
	}
	list := make([]ProxyEntry, 0, 1)
	err = json.Unmarshal([]byte(Gui.Settings.ProxyList), &list)
	if err != nil {
		return nil, err
	}
	plist := make([]*ProxyEntry, 0, len(list))
	for _, item := range list {
		entry := ProxyEntry{}
		entry = item
		var err error
		entry.Password, err = p.Crypt.DecryptPassword(string(pass), item.Password)
		if err != nil {
			return nil, err
		}
		if len(entry.KeyFileContent) > 0 {
			entry.KeyFileContent, err = p.Crypt.Decrypt([]byte(pass), item.KeyFileContent)
			if err != nil {
				return nil, err
			}
		}
		//		entry.KeyFileReader = util.ReadKeyFile
		//		entry.HostFileReader = util.ReadKeyFile
		plist = append(plist, &entry)
	}
	return plist, nil
}

func (p *Preferences) GetNumberOfProxies() (int, error) {
	list := make([]ProxyEntry, 0, 1)
	err := json.Unmarshal([]byte(Gui.Settings.ProxyList), &list)
	if err != nil {
		return 0, err
	}
	return len(list), nil
}
