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
	"embed"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"sshproxy/httpsocksserver"
	"sshproxy/socksserver"
	"sshproxy/sshtheme"
	"sshproxy/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUI struct {
	App          fyne.App
	MainWindow   fyne.Window
	Toolbar      *widget.Toolbar
	IsDesktop    bool
	Icon         *fyne.StaticResource
	FyneSettings fyne.Settings
	Settings     *Preferences
	Theme        *sshtheme.SshTheme

	toolToggleThema  *widget.ToolbarAction
	toolInfo         *widget.ToolbarAction
	toolSettings     *widget.ToolbarAction
	toolStart        *widget.ToolbarAction
	toolStop         *widget.ToolbarAction
	toolChangePasswd *widget.ToolbarAction
	toolUpdate *widget.ToolbarAction

	Scroll *container.Scroll

	Led_red_on     *fyne.StaticResource
	Led_red_off    *fyne.StaticResource
	Led_green_on   *fyne.StaticResource
	Led_green_off  *fyne.StaticResource
	Led_gray_on    *fyne.StaticResource
	Led_gray_off   *fyne.StaticResource
	Led_yellow_on  *fyne.StaticResource
	Led_yellow_off *fyne.StaticResource

	CardContainer *fyne.Container
	cards         []*SshCard
	cardMap       map[string]*SshCard

	statusTicker *time.Ticker
	statusCancel chan bool

	MasterPassword string
	running        bool
}

//go:embed assets/*
var assets embed.FS

var Gui = GUI{}

func forceLanguage() {
	if *Flags.language == "" {
		return
	}
	// Hack. Ongoing discussion in https://github.com/fyne-io/fyne/issues/5333
	lcontent, err := assets.ReadFile("assets/lang/" + *Flags.language + ".json")
	if err != nil {
		return
	}
	lang.AddTranslationsForLocale(lcontent, lang.SystemLocale())
}

type FlagsType struct {
	language *string
}

var Flags FlagsType

type Datas struct {
	datas        []*ProxyEntry
	socksServers []*socksserver.SocksServer
	httpServers  []*httpsocksserver.HttpSocksServer
	lock         sync.RWMutex
}

var Data Datas

func main() {
	Flags.language = flag.String("l", "", "language (en, de ....)")
	flag.Parse()

	loadTranslations(assets, "assets/lang")
	forceLanguage()

	//  go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	Gui.App = app.NewWithID("com.bytemystery.sshsocks")
	loadIcons()
	loadPreferences()
	Gui.FyneSettings = Gui.App.Settings()
	var tv fyne.ThemeVariant
	switch Gui.Settings.ThemeVariant {
	case -1:
		tv = fyne.CurrentApp().Settings().ThemeVariant()
	case 0:
		tv = theme.VariantDark
	case 1:
		tv = theme.VariantLight
	}

	// tv = theme.Variant

	Gui.Theme = sshtheme.NewSshTheme(tv)
	Gui.App.Settings().SetTheme(Gui.Theme)

	if _, ok := Gui.App.(desktop.App); ok {
		Gui.IsDesktop = true
	}
	Gui.MainWindow = Gui.App.NewWindow("SshProxy")
	Gui.MainWindow.SetIcon(Gui.Icon)

	Gui.toolToggleThema = widget.NewToolbarAction(theme.BrokenImageIcon(), func() {
		if Gui.Theme.GetVariant() == theme.VariantDark {
			Gui.Theme.SetVariant(theme.VariantLight)
		} else {
			Gui.Theme.SetVariant(theme.VariantDark)
		}
		Gui.Settings.ThemeVariant = int(Gui.Theme.GetVariant())
		Gui.Settings.Store()
		Gui.App.Settings().SetTheme(Gui.Theme)
		updateTheme()
	})

	Gui.toolInfo = widget.NewToolbarAction(theme.InfoIcon(), func() {
		showInfoDialog()
	})

	Gui.toolSettings = widget.NewToolbarAction(theme.SettingsIcon(), func() {
		ShowSettingsDialog()
	})

	Gui.toolStart = widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
		StartProxy()
	})
	Gui.toolStop = widget.NewToolbarAction(theme.MediaStopIcon(), func() {
		StopProxy()
	})
	Gui.toolChangePasswd = widget.NewToolbarAction(theme.AccountIcon(), func() {
		ChangePassword()
	})
	Gui.toolUpdate = widget.NewToolbarAction(theme.HelpIcon(), func() {
		CheckForUpdate(false)
	})

	Gui.Toolbar = widget.NewToolbar(Gui.toolToggleThema, widget.NewToolbarSeparator(),
		Gui.toolStart, Gui.toolStop, widget.NewToolbarSeparator(), Gui.toolSettings,
		widget.NewToolbarSeparator(), Gui.toolChangePasswd,
		widget.NewToolbarSpacer(), Gui.toolUpdate, widget.NewToolbarSeparator(), Gui.toolInfo)

	scaling := theme.Size("text") / 14.0

	Gui.CardContainer = container.NewGridWrap(fyne.NewSize(300*scaling, 350*scaling))

	Gui.Scroll = container.NewScroll(Gui.CardContainer)

	Gui.MainWindow.SetContent(container.NewBorder(Gui.Toolbar, nil, nil, nil, Gui.Scroll))

	Gui.MainWindow.Resize(fyne.NewSize(310*scaling, 395*scaling))
	Gui.MainWindow.CenterOnScreen()

	showPasswordDialog(func(pass string) {
		pass2, err := Gui.Settings.Crypt.EncryptPassword(InternPassword, pass)
		if err != nil {
			return
		}
		Gui.MasterPassword = string(pass2)
		if !CheckMasterKey() {
			dia := dialog.NewError(errors.New(lang.X("msg.masterpassword_wrong", "Masterpassword is wrong !!")), Gui.MainWindow)
			dia.SetOnClosed(func() {
				LogOut()
			})
			dia.Show()
		} else {
			if Gui.Settings.FirstStart {
				Gui.Settings.FirstStart = false
				Gui.Settings.Store()
			}
			if Gui.Settings.AutoUpdateCheck {
				CheckForUpdate(true)
			}
			UpdateCards()
		}
	}, func() {
		LogOut()
	}, Gui.Settings.FirstStart)

	fyne.CurrentApp().Settings().AddListener(func(settings fyne.Settings) {
		updateTheme()
	})

	Gui.App.Lifecycle().SetOnExitedForeground(func() {
		// if fyne.CurrentDevice().IsMobile() {
	})
	Gui.App.Lifecycle().SetOnEnteredForeground(func() {
	})

	defer CancelStatusTimer()
	defer StopProxy()

	Gui.MainWindow.Show()
	Gui.App.Run()
}

func updateTheme() {
}

func StartStatusTimer() {
	if Gui.statusTicker == nil {
		Gui.statusTicker = time.NewTicker(time.Duration(Gui.Settings.UIUpdateTime) * time.Millisecond)
		Gui.statusCancel = make(chan bool, 2)
		go UpdateStatus(Gui.statusTicker.C, Gui.statusCancel)
	} else {
		Gui.statusTicker.Reset(time.Duration(Gui.Settings.UIUpdateTime) * time.Millisecond)
	}
}

func StopStatusTimer() {
	if Gui.statusTicker != nil {
		Gui.statusTicker.Stop()
	}
}

func CancelStatusTimer() {
	if Gui.statusTicker != nil {
		Gui.statusTicker.Stop()
		Gui.statusCancel <- true
		Gui.statusCancel = nil
		Gui.statusTicker = nil
	}
}

func UpdateStatus(ticker <-chan time.Time, cancel <-chan bool) {
	for {
		select {
		case <-ticker:
			// fmt.Println("Tick:", t)
			Data.lock.RLock()
			for index := range Data.datas {
				if len(Data.socksServers) > index {
					bCon, bOn := Data.socksServers[index].IsConnected()
					Gui.cards[index].SetOnOffStatus(bCon, bOn)
					Gui.cards[index].SetStatStatus(Data.socksServers[index].GetStatistic())
				} else if len(Gui.cards) > index {
					Gui.cards[index].SetOnOffStatus(false, false)
				}
			}
			Data.lock.RUnlock()
		case <-cancel:
			fmt.Println("Beende Goroutine")
			return
		}
	}
}

func UpdateToolBar() {
	if len(Data.datas) > 0 {
		if Gui.running {
			Gui.toolStart.Disable()
			Gui.toolStop.Enable()
		} else {
			Gui.toolStart.Enable()
			Gui.toolStop.Disable()
		}
	} else {
		Gui.toolStart.Disable()
		Gui.toolStop.Disable()
	}
}

func UpdateCards() {
	StopProxy()
	proxies, _ := Gui.Settings.GetProxyList()
	f := func() {
		Gui.cards = make([]*SshCard, 0, 1)
		if len(Data.datas) > 0 {
			Data.lock.RLock()
			item := Data.datas[0]
			card := NewSshCard(item.Name, item.SocksPort, item.HttpPort)
			Gui.cards = append(Gui.cards, card)
			addCards()
			Data.lock.RUnlock()
			StartProxy()
			StartStatusTimer()
			UpdateToolBar()
		}
	}

	if len(proxies) > 1 {
		list := make([]string, 0, len(proxies))
		for _, item := range proxies {
			list = append(list, item.Name)
		}
		selProxy := widget.NewSelect(list, nil)
		index := Gui.Settings.LastSelProxyIndex
		if index >= len(proxies) {
			index = 0
		}
		selProxy.SetSelectedIndex(index)
		c := container.New(layout.NewFormLayout(),
			widget.NewLabel(lang.X("selproxy.entry", "Select proxy")), selProxy)
		dia := dialog.NewCustomConfirm(lang.X("selproxy.title", "Select proxy"), lang.X("ok", "Ok"), lang.X("cancel", "Cancel"),
			c, func(ok bool) {
				if !ok {
					return
				}
				index := selProxy.SelectedIndex()
				if index < 0 {
					return
				}
				Gui.Settings.LastSelProxyIndex = index
				Gui.Settings.Store()
				Data.datas = make([]*ProxyEntry, 1, 1)
				Data.datas[0] = proxies[index]
				f()
			}, Gui.MainWindow)
		dia.Show()
		si := Gui.MainWindow.Canvas().Size()
		var windowScale float32 = 1.0
		dia.Resize(fyne.NewSize(si.Width*windowScale, dia.MinSize().Height))
	} else {
		Data.datas = make([]*ProxyEntry, 1, 1)
		Data.datas[0] = proxies[0]
		f()
	}
}

func addCards() {
	Gui.cardMap = make(map[string]*SshCard, len(Gui.cards))
	Gui.CardContainer.RemoveAll()
	for _, item := range Gui.cards {
		Gui.CardContainer.Add(item.card)
		Gui.cardMap[item.title.GetText()] = item
	}
}

func StartProxy() {
	StopProxy()
	Data.lock.Lock()
	defer Data.lock.Unlock()
	Data.socksServers = make([]*socksserver.SocksServer, 0, len(Data.datas))
	for _, item := range Data.datas {
		s := socksserver.NewSocksServer(item.Server, item.SocksPort)
		Data.socksServers = append(Data.socksServers, s)
		s.Start()
	}
	Data.httpServers = make([]*httpsocksserver.HttpSocksServer, 0, len(Data.datas))
	for _, item := range Data.datas {
		h, _ := httpsocksserver.NewHttpSocksServer("0.0.0.0", item.SocksPort, item.HttpPort)
		Data.httpServers = append(Data.httpServers, h)
		h.Start()
	}
	Gui.running = true
	UpdateToolBar()
}

func StopProxy() {
	Data.lock.Lock()
	defer Data.lock.Unlock()
	for _, item := range Data.httpServers {
		if item != nil {
			item.Stop()
		}
	}
	for _, item := range Data.socksServers {
		if item != nil {
			item.Stop()
		}
	}
	Data.httpServers = nil
	Data.socksServers = nil
	Gui.running = false
	UpdateToolBar()
}

func CheckForUpdate(notify bool) {
	if notify {
		now := time.Now().Unix()
		if now-Gui.Settings.LastUpdatecheck < int64(Gui.Settings.UpdateCheckInterval)*3600 {
			return
		}
	}
	go func() {
		m := Gui.App.Metadata()
		type Version struct {
			maj   int
			min   int
			patch int
		}
		thisVersion := Version{}
		gitVersion := Version{}
		web, newVer, err := util.CheckForUpdate()
		if err != nil {
			return
		}
		n, err := fmt.Sscanf(m.Version, "%d.%d.%d", &thisVersion.maj, &thisVersion.min, &thisVersion.patch)
		if n != 3 || err != nil {
			return
		}
		n, err = fmt.Sscanf(newVer, "v%d.%d.%d", &gitVersion.maj, &gitVersion.min, &gitVersion.patch)
		if n != 3 || err != nil {
			return
		}
		if thisVersion.maj < gitVersion.maj || (thisVersion.maj == gitVersion.maj && thisVersion.min < gitVersion.min) ||
			(thisVersion.maj == gitVersion.maj && thisVersion.min == gitVersion.min && thisVersion.patch < gitVersion.patch) {
			link, err := url.Parse(web)
			if err != nil {
				return
			}
			fyne.Do(func() {
				if notify {
					SendNotification(lang.X("update.notify.title", "New version"), fmt.Sprintf(lang.X("update.notify.msg", "New version %s is available"), newVer))
					Gui.Settings.LastUpdatecheck = time.Now().Unix()
					Gui.Settings.Store()
				} else {
					msg := widget.NewHyperlinkWithStyle(fmt.Sprintf(lang.X("update.msg", "A new version %s is available !"), newVer),
						link, fyne.TextAlignCenter, fyne.TextStyle{
							Bold: true,
						})
					var dia *dialog.CustomDialog
					ok := widget.NewButton(lang.X("ok", "Ok"), func() {
						dia.Hide()
					})
					dia = dialog.NewCustomWithoutButtons(lang.X("update.title", "Update"),
						container.NewVBox(msg, util.NewVFiller(2), ok), Gui.MainWindow)
					dia.Show()
					dia.Resize(fyne.NewSize(Gui.MainWindow.Canvas().Size().Width, dia.MinSize().Height))
				}
			})
		} else {
			if !notify {
				fyne.Do(func() {
					dialog.ShowInformation(lang.X("update.title", "Update"), lang.X("update.nonew", "You are alread running the latest version."), Gui.MainWindow)
				})
			}
		}
	}()
}

func ChangePassword() {
	showPasswordDialog(func(pass string) {
		plist, err := Gui.Settings.GetProxyList()
		pass2, err := Gui.Settings.Crypt.EncryptPassword(InternPassword, pass)
		if err != nil {
			return
		}
		Gui.MasterPassword = string(pass2)
		Gui.Settings.MasterKeyTest = PREF_MASTERKEY_TEST_VALUE
		if !CheckMasterKey() {
			dia := dialog.NewError(errors.New(lang.X("msg.masterpassword_set_err", "Setting the masterpassword failed !!")), Gui.MainWindow)
			dia.Show()
		} else {
			Gui.Settings.SetProxies(plist)
			Gui.Settings.Store()
		}
	}, nil, true)
}
