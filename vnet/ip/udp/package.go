// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package udp

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
)

var packageIndex uint

func Init(v *vnet.Vnet) {
	m := &Main{}
	packageIndex = v.AddPackage("udp", m)
}

func GetMain(v *vnet.Vnet) *Main { return v.GetPackage(packageIndex).(*Main) }

type Main struct {
	vnet.Package
}

func (m *Main) FormatLayer(b []byte) (lines []string) {
	h := (*Header)(vnet.Pointer(b))
	lines = append(lines, h.String())
	return
}

func (m *Main) ParseLayer(b []byte, in *parse.Input) (n uint) {
	h := (*Header)(vnet.Pointer(b))
	h.Parse(in)
	return SizeofHeader
}