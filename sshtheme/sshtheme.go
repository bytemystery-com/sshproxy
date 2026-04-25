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

package sshtheme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const Scaling = 1.2

type SshTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func (p *SshTheme) GetVariant() fyne.ThemeVariant {
	return p.variant
}

func (p *SshTheme) SetVariant(variant fyne.ThemeVariant) {
	p.variant = variant
}

func NewSshTheme(variant fyne.ThemeVariant) *SshTheme {
	return &SshTheme{
		base:    theme.DefaultTheme(),
		variant: variant,
	}
}

func (p *SshTheme) Color(c fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	if p.variant == theme.VariantDark {
		switch c {
		case theme.ColorNameBackground:
			return color.NRGBA{25, 25, 25, 255}
		case theme.ColorNameOverlayBackground:
			return color.NRGBA{R: 50, G: 50, B: 50, A: 255}
		case theme.ColorNameInputBackground:
			return color.NRGBA{90, 90, 90, 255}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 50, G: 145, B: 245, A: 255}
		case theme.ColorNameButton:
			return color.NRGBA{R: 60, G: 60, B: 60, A: 255}
		case theme.ColorNamePrimary:
			return color.NRGBA{R: 87, G: 139, B: 255, A: 255}
			// return color.NRGBA{R: 41, G: 111, B: 246, A: 255}
		case theme.ColorNameSelection:
			return color.NRGBA{R: 132, G: 173, B: 255, A: 255}
		case theme.ColorNameError:
			// return color.NRGBA{{244, 67, 54, 255}
			return color.NRGBA{R: 255, G: 104, B: 82, A: 255}
		case theme.ColorNameForeground:
			return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
		}
	} else {
		switch c {
		case theme.ColorNameBackground:
			return color.NRGBA{240, 240, 240, 255}
		case theme.ColorNameOverlayBackground:
			return color.NRGBA{R: 230, G: 230, B: 230, A: 255}
		case theme.ColorNameInputBackground:
			return color.NRGBA{235, 235, 235, 255}
		case theme.ColorNameFocus:
			return color.NRGBA{R: 160, G: 200, B: 242, A: 255}
		case theme.ColorNameButton:
			return color.NRGBA{R: 225, G: 225, B: 225, A: 255}
		case theme.ColorNameSelection:
			return color.NRGBA{R: 191, G: 225, B: 255, A: 255}
		}
	}
	/*
		val := p.base.Color(c, p.variant)
		if c == theme.ColorNameForeground {
			fmt.Println(val)
		}
		return val
	*/
	return p.base.Color(c, p.variant)
}

func (p *SshTheme) Font(s fyne.TextStyle) fyne.Resource {
	return p.base.Font(s)
}

func (p *SshTheme) Icon(i fyne.ThemeIconName) fyne.Resource {
	return p.base.Icon(i)
}

func (p *SshTheme) Size(s fyne.ThemeSizeName) float32 {
	val := p.base.Size(s)
	switch s {
	case theme.SizeNameSubHeadingText, theme.SizeNameHeadingText, theme.SizeNameCaptionText, theme.SizeNameText:
		return val * Scaling
	}
	return val
}

func (p *SshTheme) GetSpecialColor(c string) color.Color {
	if p.variant == theme.VariantDark {
		switch c {
		case "error_overlay":
			return color.NRGBA{255, 70, 42, 45}
		case "wait_background":
			return color.NRGBA{127, 127, 127, 180}
		}
	} else {
		switch c {
		case "error_overlay":
			return color.NRGBA{255, 70, 105, 45}
		case "wait_background":
			return color.NRGBA{127, 127, 127, 180}
		}
	}
	return theme.Color(theme.ColorNameForeground)
}

func (p *SshTheme) GetSpecialSize(s string) float32 {
	switch s {
	case "category_view_text_scale":
		return 1.0
	case "entry_view_title_scale":
		return 1.70
	case "entry_view_datetime_scale":
		return 0.75
	case "entry_view_category_scale":
		return 0.75
	case "entry_view_label_scale":
		return 1.0
	case "category_view_icon_size":
		return 36
	case "entry_view_icon_size":
		return 48
	case "entry_edit_icon_size":
		return 48
	case "icon_select_icon_size":
		return 64
	case "login_icon_size":
		return 150
	case "login_space_logo_label":
		return 5
	case "login_space_field_ok":
		return 3
	}
	return 1.0
}
