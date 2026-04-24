package main

import (
	"io"
	"strconv"

	"sshproxy/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/crypto/ssh"
)

type ProxyCfgTab struct {
	name          *widget.Entry
	user          *widget.Entry
	host          *widget.Entry
	port          *widget.Entry
	socksPort     *widget.Entry
	httpPort      *widget.Entry
	pass          *widget.Entry
	keyFile       *widget.Label
	keyFileBrowse *widget.Button
	keyFileDel    *widget.Button
	hostKeyList   *widget.List
	toolItemAdd   *widget.ToolbarAction
	toolItemDel   *widget.ToolbarAction
	toolBar       *widget.Toolbar
	form          *fyne.Container
	tabItem       *container.TabItem
	hostFiles     [][]byte
	proxy         *ProxyEntry
}

func getPrivateKeyFingerPrint(key, pass []byte) (string, error) {
	signer, err := ssh.ParsePrivateKey(key)
	_, ok := err.(*ssh.PassphraseMissingError)
	if ok {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, pass)
	}
	if err == nil {
		return ssh.FingerprintSHA256(signer.PublicKey()), nil
	}
	return "", err
}

func getPublicKeyFingerPrint(key []byte) (string, error) {
	pubKey, comment, _, _, err := ssh.ParseAuthorizedKey(key)
	if err == nil {
		str := ssh.FingerprintSHA256(pubKey)
		if comment != "" {
			str += " - " + comment
		}
		return str, nil
	}
	return "", err
}

func NewProxyCfgTab(index int, proxy *ProxyEntry) *ProxyCfgTab {
	selectedHostFileIndex := -1

	cfgTab := ProxyCfgTab{
		proxy: proxy,
	}
	if proxy == nil {
		cfgTab.proxy = &ProxyEntry{}
	}
	cfgTab.name = widget.NewEntry()
	cfgTab.name.SetPlaceHolder(lang.X("settings.name_placeholder", "Name for display - for deleting an entry leave this field empty"))
	cfgTab.user = widget.NewEntry()
	cfgTab.user.SetPlaceHolder(lang.X("settings.user_placeholder", "SSH user"))
	cfgTab.host = widget.NewEntry()
	cfgTab.host.SetPlaceHolder(lang.X("settings.host_placeholder", "SSH host (xy.com)"))

	cfgTab.port = widget.NewEntry()
	cfgTab.port.SetPlaceHolder(lang.X("settings.port_placeholder", "SSH port (22)"))
	cfgTab.port.OnChanged = util.GetNumberFilter(cfgTab.port, nil)

	cfgTab.socksPort = widget.NewEntry()
	cfgTab.socksPort.SetPlaceHolder(lang.X("settings.socksport_placeholder", "SOCKS port (7777)"))
	cfgTab.socksPort.OnChanged = util.GetNumberFilter(cfgTab.socksPort, nil)

	cfgTab.httpPort = widget.NewEntry()
	cfgTab.httpPort.SetPlaceHolder(lang.X("settings.httpport_placeholder", "HTTP port (8888)"))
	cfgTab.httpPort.OnChanged = util.GetNumberFilter(cfgTab.httpPort, nil)

	cfgTab.pass = widget.NewPasswordEntry()
	cfgTab.pass.SetPlaceHolder(lang.X("settings.pass_placeholder", "SSH key - password"))
	cfgTab.keyFile = widget.NewLabel("")
	cfgTab.keyFileBrowse = widget.NewButton(lang.X("settings.browse", "Browse"), func() {
		dia := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				return
			}
			if r == nil {
				return
			}
			data, err := io.ReadAll(r)
			if err == nil {
				cfgTab.proxy.KeyFileContent = data
				go func() {
					str, err := getPrivateKeyFingerPrint(cfgTab.proxy.KeyFileContent, []byte(cfgTab.pass.Text))
					fyne.Do(func() {
						if err != nil {
							UIErrorHandler(err)
						} else {
							cfgTab.keyFile.SetText(str)
						}
					})
				}()
			}
		}, Gui.MainWindow)

		dia.SetView(dialog.ListView)
		ms := Gui.MainWindow.Canvas().Size()
		dia.Resize(fyne.NewSize(ms.Width*.8, ms.Height*.8))
		dia.Show()
	})
	cfgTab.keyFileDel = widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), func() {
		cfgTab.proxy.KeyFileContent = []byte{}
		cfgTab.keyFile.SetText("")
	})

	listLength := func() int {
		return len(cfgTab.hostFiles)
	}
	listCreate := func() fyne.CanvasObject {
		text := canvas.NewText("", theme.Color(theme.ColorNameForeground))
		text.Refresh()
		return text
	}

	listUpdate := func(id widget.ListItemID, o fyne.CanvasObject) {
		text, ok := o.(*canvas.Text)
		if !ok {
			return
		}
		str, err := getPublicKeyFingerPrint(cfgTab.hostFiles[id])
		if err != nil {
			return
		}
		text.Text = str
		text.Color = theme.Color(theme.ColorNameForeground)
		text.Refresh()
	}

	cfgTab.hostKeyList = widget.NewList(listLength, listCreate, listUpdate)
	selectedHostFileIndex = -1
	cfgTab.hostKeyList.OnSelected = func(id widget.ListItemID) {
		selectedHostFileIndex = id
	}
	cfgTab.hostKeyList.OnUnselected = func(id widget.ListItemID) {
		selectedHostFileIndex = -1
	}

	cfgTab.toolItemAdd = widget.NewToolbarAction(theme.ContentAddIcon(), func() {
		diaHost := dialog.NewFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, Gui.MainWindow)
				return
			}
			if r == nil {
				return
			}
			defer r.Close()
			data, err := io.ReadAll(r)
			_, _, _, _, err = ssh.ParseAuthorizedKey(data)
			if err != nil {
				dialog.ShowError(err, Gui.MainWindow)
				return
			}
			cfgTab.hostFiles = append(cfgTab.hostFiles, data)
			cfgTab.hostKeyList.Refresh()
		}, Gui.MainWindow)
		diaHost.SetView(dialog.ListView)
		ms := Gui.MainWindow.Canvas().Size()
		diaHost.Resize(fyne.NewSize(ms.Width*.8, ms.Height*.8))
		diaHost.Show()
	})
	cfgTab.toolItemDel = widget.NewToolbarAction(theme.ContentRemoveIcon(), func() {
		if selectedHostFileIndex >= 0 {
			cfgTab.hostFiles = append(cfgTab.hostFiles[:selectedHostFileIndex], cfgTab.hostFiles[selectedHostFileIndex+1:]...)
			cfgTab.hostKeyList.Refresh()
		}
	})
	cfgTab.toolBar = widget.NewToolbar(cfgTab.toolItemAdd, cfgTab.toolItemDel)

	cfgTab.form = container.New(layout.NewFormLayout(),
		widget.NewLabel(lang.X("settings.name", "Name")), cfgTab.name,
		widget.NewLabel(lang.X("settings.socksport", "SOCKS port")), cfgTab.socksPort,
		widget.NewLabel(lang.X("settings.httpport", "HTTP port")), cfgTab.httpPort,
		widget.NewLabel(lang.X("settings.host", "Host")), cfgTab.host,
		widget.NewLabel(lang.X("settings.port", "Port")), cfgTab.port,
		widget.NewLabel(lang.X("settings.user", "User")), cfgTab.user,
		widget.NewLabel(lang.X("settings.password", "Password")), cfgTab.pass,
		widget.NewLabel(lang.X("settings.sshkey", "SSH key")), container.NewBorder(nil, nil, nil, container.NewHBox(cfgTab.keyFileBrowse, cfgTab.keyFileDel), cfgTab.keyFile),
		widget.NewLabel(lang.X("settings.hostkeys", "Host keys")), cfgTab.toolBar,
	)
	content := container.NewBorder(cfgTab.form, nil, nil, nil, cfgTab.hostKeyList)
	cfgTab.tabItem = container.NewTabItem(strconv.Itoa(index), content)

	if proxy != nil {
		cfgTab.name.SetText(proxy.Name)
		cfgTab.host.SetText(proxy.Host)
		cfgTab.port.SetText(strconv.Itoa(proxy.Port))
		cfgTab.socksPort.SetText(strconv.Itoa(proxy.SocksPort))
		cfgTab.httpPort.SetText(strconv.Itoa(proxy.HttpPort))
		cfgTab.user.SetText(proxy.User)
		cfgTab.pass.SetText(proxy.Password)
		cfgTab.keyFile.SetText(proxy.KeyFile)
		go func() {
			str, err := getPrivateKeyFingerPrint(proxy.KeyFileContent, []byte(cfgTab.pass.Text))
			fyne.Do(func() {
				if err != nil {
					UIErrorHandler(err)
				} else {
					cfgTab.keyFile.SetText(str)
				}
			})
		}()
		cfgTab.hostFiles = proxy.HostFilesContent
	}
	return &cfgTab
}

func ShowSettingsDialog() {
	tabs := container.NewAppTabs()
	entries, _ := Gui.Settings.GetProxyList()
	anzahl := len(entries) + 1

	cfgTabList := make([]*ProxyCfgTab, 0, anzahl)

	for i := range anzahl {
		var proxy *ProxyEntry
		if i < len(entries) {
			proxy = entries[i]
		}
		cfg := NewProxyCfgTab(i+1, proxy)
		cfgTabList = append(cfgTabList, cfg)
		tabs.Append(cfg.tabItem)
	}
	oldSize := Gui.MainWindow.Canvas().Size()
	Gui.MainWindow.Resize(fyne.NewSize(950, 700))
	Gui.MainWindow.CenterOnScreen()
	dia := dialog.NewCustomConfirm(lang.X("settigs.cation", "Settings"),
		lang.X("ok", "Ok"), lang.X("cancel", "Cancel"), container.NewScroll(tabs), func(ok bool) {
			// to do
			Gui.MainWindow.Resize(oldSize)
			Gui.MainWindow.CenterOnScreen()
			if !ok {
				return
			}
			pList := make([]*ProxyEntry, 0, len(tabs.Items))
			for i := range len(tabs.Items) {
				cfgTab := cfgTabList[i]
				if cfgTab.name.Text != "" {
					proxy := ProxyEntry{}
					proxy.Name = cfgTab.name.Text
					proxy.Host = cfgTab.host.Text
					x, err := strconv.Atoi(cfgTab.port.Text)
					if err == nil {
						proxy.Port = x
					}
					x, err = strconv.Atoi(cfgTab.socksPort.Text)
					if err == nil {
						if x <= 0 {
							x = 7777
						}
						proxy.SocksPort = x
					} else {
						proxy.SocksPort = 7777
					}
					x, err = strconv.Atoi(cfgTab.httpPort.Text)
					if err == nil {
						if x <= 0 {
							x = 8888
						}
						proxy.HttpPort = x
					} else {
						proxy.HttpPort = 8888
					}
					proxy.User = cfgTab.user.Text
					proxy.Password = cfgTab.pass.Text
					proxy.KeyFileContent = cfgTab.proxy.KeyFileContent
					proxy.HostFilesContent = cfgTab.hostFiles
					pList = append(pList, &proxy)
				}
			}
			Gui.Settings.SetProxies(pList)
			Gui.Settings.Store()
			UpdateCards()
		}, Gui.MainWindow)

	dia.Show()
	//	Gui.MainWindow.Canvas().Focus(user)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 1.0
	dia.Resize(fyne.NewSize(si.Width*windowScale, si.Height*windowScale))
}
