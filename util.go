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
	"fmt"
	"net/url"
	"runtime"
	"runtime/debug"

	"sshproxy/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func showInfoDialog() {
	vgo := runtime.Version()[2:]
	vfyne := ""
	os := runtime.GOOS
	arch := runtime.GOARCH
	info, _ := debug.ReadBuildInfo()
	for _, dep := range info.Deps {
		if dep.Path == "fyne.io/fyne/v2" {
			vfyne = dep.Version[1:]
		}
	}
	s := fyne.CurrentApp().Settings()
	t := Gui.Theme.GetVariant()
	thema := ""
	b := s.BuildType()
	_ = b
	switch t {
	case theme.VariantDark:
		thema = lang.X("info.thema_dark", "Dark")
	case theme.VariantLight:
		thema = lang.X("info.thema_light", "Light")
	default:
		thema = lang.X("info.thema_unknown", "Unknown")
	}

	build := ""
	switch b {
	case fyne.BuildStandard:
		build = lang.X("info.build_standard", "Standard")
	case fyne.BuildDebug:
		build = lang.X("info.build_debug", "Debug")
	case fyne.BuildRelease:
		build = lang.X("info.build_release", "Release")
	default:
		build = lang.X("info.build_unknown", "Unknown")
	}

	m := Gui.App.Metadata()
	v := fmt.Sprintf("%s (%d)", m.Version, m.Build)
	n := m.Name
	if n == "" {
		n = "Bilderwörterbuch"
	}
	tsStr := ""
	ts := m.Custom["buildts"]
	if ts != "" {
		tsStr = "Build: " + ts + "\n"
	}
	wSize := Gui.MainWindow.Canvas().Size()

	msg := fmt.Sprintf(lang.X("info.msg", "%s\nVersion: %s  \n%sAuthor: Reiner Pröls\n\nGo version: %s\nFyne version: %s\nBuild: %s\nThema: %s\nWindow size: %.0fx%.0f\nPlatform: %s\nArchitecture: %s"),
		n, v, tsStr, vgo, vfyne, build, thema, wSize.Width, wSize.Height, os, arch)
	dialog.ShowInformation(lang.X("info.title", "Info"), msg, Gui.MainWindow)
}

func loadPreferences() {
	Gui.Settings = NewPreferences()
}

func loadIcon(path, name string) *fyne.StaticResource {
	data, err := assets.ReadFile(path)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(name, data)
}

func loadTranslations(fs embed.FS, dir string) {
	lang.AddTranslationsFS(fs, dir)
}

func loadIcons() {
	Gui.Icon = loadIcon("assets/icons/icon.png", "icon")
	Gui.App.SetIcon(Gui.Icon)

	Gui.Led_green_on = loadIcon("assets/icons/led_green_on.png", "led_green_on")
	Gui.Led_green_off = loadIcon("assets/icons/led_green_off.png", "led_green_off")
	Gui.Led_red_on = loadIcon("assets/icons/led_red_on.png", "led_red_on")
	Gui.Led_red_off = loadIcon("assets/icons/led_red_off.png", "led_red_off")
	Gui.Led_gray_on = loadIcon("assets/icons/led_x_on.png", "led_x_on")
	Gui.Led_gray_off = loadIcon("assets/icons/led_x_off.png", "led_x_off")
	Gui.Led_yellow_on = loadIcon("assets/icons/led_yellow_on.png", "led_yellow_on")
	Gui.Led_yellow_off = loadIcon("assets/icons/led_yellow_off.png", "led_yellow_off")

	loadIconsForTheme()
}

func loadIconsForTheme() {
	dir := ""
	switch fyne.CurrentApp().Settings().ThemeVariant() {
	case theme.VariantDark:
		dir = "dark"
	case theme.VariantLight:
		dir = "light"
	default:
		dir = "light"
	}
	/*	Gui.IconImport = loadIcon("assets/icons/"+dir+"/import.png", "import")
		Gui.IconExport = loadIcon("assets/icons/"+dir+"/export.png", "export")
	*/
	_ = dir
}

func SendNotification(title, msg string) {
	fyne.Do(func() {
		n := fyne.NewNotification(title, msg)
		Gui.App.SendNotification(n)
	})
}

func doHelp() {
	u := url.URL{
		Scheme: "https",
		Host:   "github.com",
		Path:   "/bytemystery-com/passwordsafe",
	}
	Gui.App.OpenURL(&u)
}

func UIErrorHandler(err error) {
	UIErrorHandlerWithMessage(err, "")
}

func UIErrorHandlerWithMessage(err error, msg string) {
	fyne.Do(func() {
		if msg != "" {
			if msg[len(msg)-1] != '\n' {
				msg += "\n"
			}
			err = errors.Join(errors.New(msg), err)
		}
		dialog.ShowError(err, Gui.MainWindow)
	})
}

func showPasswordDialog(fOk func(pass string), fCancel func(), withConfirm bool) {
	var dia *dialog.ConfirmDialog

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder(lang.X("masterpasswd.dialog.passwdplaceholder", "Master password"))
	passEntry.OnSubmitted = func(string) {
		dia.Confirm()
	}

	passEntryConfirm := widget.NewPasswordEntry()
	passEntryConfirm.SetPlaceHolder(lang.X("masterpasswd.dialog.confirm.passwdplaceholder", "Retype master password"))
	passEntryConfirm.OnSubmitted = func(string) {
		dia.Confirm()
	}

	validator := func(str string) error {
		if len(passEntry.Text) < 3 {
			return errors.New("Passwords is too short")
		}
		return nil
	}

	validatorConfirm := func(str string) error {
		if passEntryConfirm.Text != passEntry.Text {
			return errors.New("Passwords does not match")
		}
		return nil
	}

	if withConfirm {
		passEntry.Validator = validator
		passEntryConfirm.Validator = validatorConfirm
	}

	confirm := func(confirm bool) {
		if confirm {
			err := passEntry.Validate()
			if err != nil {
				dia.Show()
				return
			}
			if withConfirm {
				err := passEntryConfirm.Validate()
				if err != nil {
					dia.Show()
					return
				}
			}

			fOk(passEntry.Text)
		} else {
			if fCancel != nil {
				fCancel()
			}
		}
	}
	var c *fyne.Container
	if withConfirm {
		c = container.NewVBox(passEntry, widget.NewLabel(lang.X("masterpasswd.dialog.confirm", "Confirm password")), passEntryConfirm, util.NewVFiller(1.0))
	} else {
		c = container.NewVBox(passEntry, util.NewVFiller(1.0))
	}
	t := ""
	if withConfirm {
		t = lang.X("masterpasswd.dialog.title.new", "New master password")
	} else {
		t = lang.X("masterpasswd.dialog.title", "Master password")
	}

	dia = dialog.NewCustomConfirm(t, lang.X("ok", "Ok"), lang.X("cancel", "Cancel"),
		c, confirm, Gui.MainWindow)
	dia.Show()

	Gui.MainWindow.Canvas().Focus(passEntry)
	si := Gui.MainWindow.Canvas().Size()
	var windowScale float32 = 1.0
	dia.Resize(fyne.NewSize(si.Width*windowScale, dia.MinSize().Height))
}

func CheckMasterKey() bool {
	pass, err := Gui.Settings.Crypt.DecryptPassword(InternPassword, Gui.MasterPassword)
	if err != nil {
		return false
	}
	if Gui.Settings.MasterKeyTest == PREF_MASTERKEY_TEST_VALUE {
		x, err := Gui.Settings.Crypt.EncryptPassword(string(pass), PREF_MASTERKEY_TEST_VALUE)
		if err != nil {
			return false
		}
		Gui.Settings.MasterKeyTest = x
		Gui.Settings.Store()
		return true
	}
	t, err := Gui.Settings.Crypt.DecryptPassword(string(pass), Gui.Settings.MasterKeyTest)
	if err != nil {
		return false
	}
	if string(t) == PREF_MASTERKEY_TEST_VALUE {
		return true
	}
	return false
}

func LogOut() {
	showPasswordDialog(func(pass string) {
		pass2, err := Gui.Settings.Crypt.EncryptPassword(InternPassword, pass)
		if err != nil {
			return
		}
		Gui.MasterPassword = string(pass2)
		pass = ""
		if !CheckMasterKey() {
			dia := dialog.NewError(errors.New(lang.X("msg.masterpassword_wrong", "Masterpassword is wrong !!")), Gui.MainWindow)
			dia.Show()
			dia.SetOnClosed(func() {
				LogOut()
			})
		} else {
			UpdateToolBar()
		}
	}, func() {
		LogOut()
	}, false)
}
