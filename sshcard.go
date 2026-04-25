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
	"fmt"
	"strconv"
	"time"

	"sshproxy/util"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bytemystery-com/colorlabel"
)

var (
	labelTotalColor       = theme.ColorNameError
	labelReadColor        = theme.ColorNameForeground
	labelWriteColor       = theme.ColorNameForeground
	labelReconnectColor   = theme.ColorNameForeground
	labelSocksPortColor   = theme.ColorNameForeground
	labelHttpPortColor    = theme.ColorNameForeground
	labelLastConnectColor = theme.ColorNameForeground
	labelScale            = 1.2
	labelBold             = true
	labelMonospace        = false

	displayTotalColor       = theme.ColorNameError
	displayReadColor        = theme.ColorNameForeground
	displayWriteColor       = theme.ColorNameForeground
	displayReconnectColor   = theme.ColorNameForeground
	displaySocksPortColor   = theme.ColorNamePrimary
	displayHttpPortColor    = theme.ColorNamePrimary
	displayLastConnectColor = theme.ColorNamePrimary
	displayScale            = 1.2
	displayBold             = true
	displayMonospace        = false

	unitTotalColor       = theme.ColorNameError
	unitReadColor        = theme.ColorNameForeground
	unitWriteColor       = theme.ColorNameForeground
	unitLastConnectColor = theme.ColorNameForeground
	unitScale            = 1.2
	unitBold             = true
	unitMonospace        = false
)

type SshCard struct {
	title            *colorlabel.ColorLabel
	labelTotal       *colorlabel.ColorLabel
	labelRead        *colorlabel.ColorLabel
	labelWrite       *colorlabel.ColorLabel
	labelReconnect   *colorlabel.ColorLabel
	labelSocksPort   *colorlabel.ColorLabel
	labelHttpPort    *colorlabel.ColorLabel
	labelLastConnect *colorlabel.ColorLabel

	displaySocksPort   *colorlabel.ColorLabel
	displayHttpPort    *colorlabel.ColorLabel
	displayTotal       *colorlabel.ColorLabel
	displayRead        *colorlabel.ColorLabel
	displayWrite       *colorlabel.ColorLabel
	displayReconnect   *colorlabel.ColorLabel
	displayLastConnect *colorlabel.ColorLabel

	unitTotal       *colorlabel.ColorLabel
	unitRead        *colorlabel.ColorLabel
	unitWrite       *colorlabel.ColorLabel
	unitLastConnect *colorlabel.ColorLabel

	icon   *widget.Icon
	card   *widget.Card
	online bool
	on     bool
}

func NewSshCard(title string, socksPort, httpPort int) *SshCard {
	//	textWidth := util.GetDefaultTextWidth("X")

	p := SshCard{
		title: colorlabel.NewColorLabel(title, theme.ColorNamePrimary, nil, 2.0),
		icon:  widget.NewIcon(Gui.Led_gray_off),
	}
	AdjustImage(p.icon)

	header := container.NewHBox(p.title, layout.NewSpacer(), p.icon)
	labelStyle := fyne.TextStyle{
		Bold:      labelBold,
		Monospace: labelMonospace,
	}
	displayStyle := fyne.TextStyle{
		Bold:      displayBold,
		Monospace: displayMonospace,
	}
	unitStyle := fyne.TextStyle{
		Bold:      unitBold,
		Monospace: unitMonospace,
	}
	p.labelSocksPort = colorlabel.NewColorLabel(lang.X("card.socksport", "SOCKS port"), labelSocksPortColor, nil, float32(labelScale))
	p.labelSocksPort.SetTextStyle(&labelStyle)
	p.labelHttpPort = colorlabel.NewColorLabel(lang.X("card.httpport", "HTTP port"), labelHttpPortColor, nil, float32(labelScale))
	p.labelHttpPort.SetTextStyle(&labelStyle)

	p.labelTotal = colorlabel.NewColorLabel(lang.X("card.total", "Total"), labelTotalColor, nil, float32(labelScale))
	p.labelTotal.SetTextStyle(&labelStyle)
	p.labelRead = colorlabel.NewColorLabel(lang.X("card.read", "Read"), labelReadColor, nil, float32(labelScale))
	p.labelRead.SetTextStyle(&labelStyle)
	p.labelWrite = colorlabel.NewColorLabel(lang.X("card.write", "Write"), labelWriteColor, nil, float32(labelScale))
	p.labelWrite.SetTextStyle(&labelStyle)

	p.labelLastConnect = colorlabel.NewColorLabel(lang.X("card.lastConnect", "Online"), labelLastConnectColor, nil, float32(labelScale))
	p.labelLastConnect.SetTextStyle(&labelStyle)

	p.displaySocksPort = colorlabel.NewColorLabel(strconv.Itoa(socksPort), displaySocksPortColor, nil, float32(displayScale))
	p.displaySocksPort.SetTextStyle(&displayStyle)
	p.displayHttpPort = colorlabel.NewColorLabel(strconv.Itoa(httpPort), displayHttpPortColor, nil, float32(displayScale))
	p.displayHttpPort.SetTextStyle(&displayStyle)

	p.displayTotal = colorlabel.NewColorLabel("", displayTotalColor, nil, float32(displayScale))
	p.displayTotal.SetTextStyle(&displayStyle)
	p.displayRead = colorlabel.NewColorLabel("", displayReadColor, nil, float32(displayScale))
	p.displayRead.SetTextStyle(&displayStyle)
	p.displayWrite = colorlabel.NewColorLabel("", displayWriteColor, nil, float32(displayScale))
	p.displayWrite.SetTextStyle(&displayStyle)
	p.displayLastConnect = colorlabel.NewColorLabel("", displayLastConnectColor, nil, float32(displayScale))
	p.displayLastConnect.SetTextStyle(&displayStyle)

	p.unitTotal = colorlabel.NewColorLabel(lang.X("card.kbyte", "kByte"), unitTotalColor, nil, float32(unitScale))
	p.unitTotal.SetTextStyle(&unitStyle)
	p.unitRead = colorlabel.NewColorLabel(lang.X("card.kbyte", "kByte"), unitReadColor, nil, float32(unitScale))
	p.unitRead.SetTextStyle(&unitStyle)
	p.unitWrite = colorlabel.NewColorLabel(lang.X("card.kbyte", "kByte"), unitWriteColor, nil, float32(unitScale))
	p.unitWrite.SetTextStyle(&unitStyle)
	p.unitLastConnect = colorlabel.NewColorLabel(lang.X("card.sec", "sec"), unitLastConnectColor, nil, float32(unitScale))
	p.unitLastConnect.SetTextStyle(&unitStyle)

	var content *fyne.Container
	labelSize := util.GetDefaultTextSize("XXXXXXXXXXXXXX")
	fieldSize := util.GetDefaultTextSize("XXXXXXXXXX")
	unitSize := util.GetDefaultTextSize("XXX")
	fieldSize.Height *= 1.6
	labelSize.Height *= 1.6
	unitSize.Height *= 1.6

	p.labelReconnect = colorlabel.NewColorLabel(lang.X("card.reconnects", "Reconnects"), labelReconnectColor, nil, float32(labelScale))
	p.labelReconnect.SetTextStyle(&labelStyle)
	p.displayReconnect = colorlabel.NewColorLabel("", displayReconnectColor, nil, float32(displayScale))
	p.displayReconnect.SetTextStyle(&displayStyle)

	content = container.New(layout.NewFormLayout(),
		container.NewGridWrap(labelSize, p.labelSocksPort), container.NewHBox(container.NewGridWrap(fieldSize, p.displaySocksPort)),
		container.NewGridWrap(labelSize, p.labelHttpPort), container.NewHBox(container.NewGridWrap(fieldSize, p.displayHttpPort)),
		container.NewGridWrap(labelSize, p.labelTotal), container.NewHBox(container.NewGridWrap(fieldSize, p.displayTotal), container.NewGridWrap(unitSize, p.unitTotal)),
		container.NewGridWrap(labelSize, p.labelRead), container.NewHBox(container.NewGridWrap(fieldSize, p.displayRead), container.NewGridWrap(unitSize, p.unitRead)),
		container.NewGridWrap(labelSize, p.labelWrite), container.NewHBox(container.NewGridWrap(fieldSize, p.displayWrite), container.NewGridWrap(unitSize, p.unitWrite)),
		container.NewGridWrap(labelSize, p.labelReconnect), container.NewGridWrap(fieldSize, container.NewGridWrap(unitSize, p.displayReconnect)),
		container.NewGridWrap(labelSize, p.labelLastConnect), container.NewHBox(container.NewGridWrap(fieldSize, p.displayLastConnect), container.NewGridWrap(unitSize, p.unitLastConnect)),
	)

	content = container.NewVBox(header, content)
	fillerSize := float32(16)
	c := container.NewBorder(util.NewFiller(fillerSize, fillerSize), util.NewFiller(fillerSize, fillerSize), util.NewFiller(fillerSize, fillerSize), util.NewFiller(fillerSize, fillerSize),
		content)
	p.card = widget.NewCard("", "", c)
	return &p
}

func (c *SshCard) SetTitle(str string) {
	c.title.SetText(str)
}

func (c *SshCard) SetOnOffStatus(online bool, on bool) {
	fyne.Do(func() {
		c.online = online
		c.on = on
		c.icon.Resource = c.getLedIcon(online, on)
		AdjustImage(c.icon)
		if !online {
			c.displayLastConnect.SetText("---")
			c.unitLastConnect.SetText("---")
		}
	})
}

/*
func (c *SshCard) SetStatus(r uint64, w uint64, reconnects uint32, online, off bool) {
	fyne.Do(func() {
		c.SetOnOffStatus(online, off)

		val, unit := formatBytesDisplay(r)
		c.displayRead.SetText(val)
		c.unitRead.SetText(unit)

		val, unit = formatBytesDisplay(w)
		c.displayWrite.SetText(val)
		c.unitWrite.SetText(unit)

		val, unit = formatBytesDisplay(r + w)
		c.displayTotal.SetText(val)
		c.unitTotal.SetText(unit)

		c.displayReconnect.SetText(fmt.Sprintf("%d", reconnects))
	})
}
*/

func (c *SshCard) SetStatStatus(r uint64, w uint64, reconnects uint64, lastConnect time.Time) {
	fyne.Do(func() {
		val, unit := formatBytesDisplay(r)
		c.displayRead.SetText(val)
		c.unitRead.SetText(unit)

		val, unit = formatBytesDisplay(w)
		c.displayWrite.SetText(val)
		c.unitWrite.SetText(unit)

		val, unit = formatBytesDisplay(r + w)
		c.displayTotal.SetText(val)
		c.unitTotal.SetText(unit)

		if !lastConnect.IsZero() {
			dur := time.Since(lastConnect).Seconds()
			val, unit = formatTimeDisplay(dur)
			c.displayLastConnect.SetText(val)
			c.unitLastConnect.SetText(unit)
		}

		c.displayReconnect.SetText(fmt.Sprintf("%d", reconnects))
	})
}

func formatTimeDisplay(v float64) (string, string) {
	if v < 60 {
		return fmt.Sprintf("%.0f", v), lang.X("card.sec", "sec")
	} else if v < 60*60 {
		x := int(v / 60)
		return fmt.Sprintf("%.d:%02d", x, int(v-float64(60*x))), lang.X("card.sec", "m:s")
	} else if v < 60*60*24 {
		x1 := int(v / (60 * 60))
		x2 := int((v - float64(60*60*x1)) / 60)
		return fmt.Sprintf("%d:%02.d:%02d", x1, x2, int(v-float64(60*60*x1)-float64(60*x2))), lang.X("card.min", "h:m:s")
	} else {
		return fmt.Sprintf("%.2f", v/(60*60*24)), lang.X("card.hour", "h:m:s")
	}
}

func formatBytesDisplay(v uint64) (string, string) {
	val := float64(v)
	if val < 1000*1000 {
		return fmt.Sprintf("%.0f", float64(val/1000.0)), lang.X("card.kbyte", "kByte")
	} else if v < 1000*1000*1000 {
		return fmt.Sprintf("%.1f", float64(val/(1000.0*1000.0))), lang.X("card.mbyte", "MByte")
	} else if v < 1000*1000*1000*1000 {
		return fmt.Sprintf("%.2f", float64(val/(1000.0*1000.0*1000.0))), lang.X("card.gbyte", "GByte")
	} else {
		return fmt.Sprintf("%.2f", float64(val/(1000.0*1000.0*1000.0))), lang.X("card.tbyte", "TByte")
	}
}

func AdjustImage(img *widget.Icon) {
	img.Refresh()
}

func (c *SshCard) getLedIcon(online, on bool) *fyne.StaticResource {
	if online {
		return Gui.Led_green_on
	} else {
		if on {
			return Gui.Led_red_on
		}
		return Gui.Led_gray_on
	}
}
