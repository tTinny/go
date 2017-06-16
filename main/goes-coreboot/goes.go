// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/bang"
	"github.com/platinasystems/go/goes/cmd/boot"
	"github.com/platinasystems/go/goes/cmd/cat"
	"github.com/platinasystems/go/goes/cmd/cd"
	"github.com/platinasystems/go/goes/cmd/chmod"
	"github.com/platinasystems/go/goes/cmd/cli"
	"github.com/platinasystems/go/goes/cmd/cmdline"
	"github.com/platinasystems/go/goes/cmd/cp"
	"github.com/platinasystems/go/goes/cmd/daemons"
	"github.com/platinasystems/go/goes/cmd/dmesg"
	"github.com/platinasystems/go/goes/cmd/echo"
	"github.com/platinasystems/go/goes/cmd/env"
	"github.com/platinasystems/go/goes/cmd/exec"
	"github.com/platinasystems/go/goes/cmd/exit"
	"github.com/platinasystems/go/goes/cmd/export"
	"github.com/platinasystems/go/goes/cmd/femtocom"
	"github.com/platinasystems/go/goes/cmd/hdel"
	"github.com/platinasystems/go/goes/cmd/hdelta"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/hexists"
	"github.com/platinasystems/go/goes/cmd/hget"
	"github.com/platinasystems/go/goes/cmd/hgetall"
	"github.com/platinasystems/go/goes/cmd/hkeys"
	"github.com/platinasystems/go/goes/cmd/hset"
	"github.com/platinasystems/go/goes/cmd/iminfo"
	"github.com/platinasystems/go/goes/cmd/insmod"
	"github.com/platinasystems/go/goes/cmd/install"
	"github.com/platinasystems/go/goes/cmd/kexec"
	"github.com/platinasystems/go/goes/cmd/keys"
	"github.com/platinasystems/go/goes/cmd/kill"
	"github.com/platinasystems/go/goes/cmd/ln"
	"github.com/platinasystems/go/goes/cmd/log"
	"github.com/platinasystems/go/goes/cmd/ls"
	"github.com/platinasystems/go/goes/cmd/lsmod"
	"github.com/platinasystems/go/goes/cmd/mkdir"
	"github.com/platinasystems/go/goes/cmd/mknod"
	"github.com/platinasystems/go/goes/cmd/mount"
	"github.com/platinasystems/go/goes/cmd/ping"
	"github.com/platinasystems/go/goes/cmd/ps"
	"github.com/platinasystems/go/goes/cmd/pwd"
	"github.com/platinasystems/go/goes/cmd/reboot"
	"github.com/platinasystems/go/goes/cmd/redisd"
	"github.com/platinasystems/go/goes/cmd/reload"
	"github.com/platinasystems/go/goes/cmd/restart"
	"github.com/platinasystems/go/goes/cmd/rm"
	"github.com/platinasystems/go/goes/cmd/rmmod"
	"github.com/platinasystems/go/goes/cmd/show_commands"
	"github.com/platinasystems/go/goes/cmd/show_packages"
	"github.com/platinasystems/go/goes/cmd/slashinit"
	"github.com/platinasystems/go/goes/cmd/sleep"
	"github.com/platinasystems/go/goes/cmd/source"
	"github.com/platinasystems/go/goes/cmd/start"
	"github.com/platinasystems/go/goes/cmd/stop"
	"github.com/platinasystems/go/goes/cmd/stty"
	"github.com/platinasystems/go/goes/cmd/subscribe"
	"github.com/platinasystems/go/goes/cmd/sync"
	"github.com/platinasystems/go/goes/cmd/umount"
	"github.com/platinasystems/go/goes/cmd/uninstall"
	"github.com/platinasystems/go/goes/cmd/wget"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "goes-coreboot"
	Apropos = "the coreboot goes machine"
)

func Goes() *goes.Goes {
	g := goes.New(Name, "",
		lang.Alt{
			lang.EnUS: Apropos,
		},
		lang.Alt{},
	)
	g.Plot(helpers.New()...)
	g.Plot(cli.New()...)
	g.Plot(bang.New(),
		boot.New(),
		cat.New(),
		cd.New(),
		chmod.New(),
		cmdline.New(),
		cp.New(),
		daemons.New(),
		dmesg.New(),
		echo.New(),
		env.New(),
		exec.New(),
		exit.New(),
		export.New(),
		femtocom.New(),
		hdel.New(),
		hdelta.New(),
		hexists.New(),
		hget.New(),
		hgetall.New(),
		hkeys.New(),
		hset.New(),
		iminfo.New(),
		insmod.New(),
		install.New(),
		kexec.New(),
		keys.New(),
		kill.New(),
		ln.New(),
		log.New(),
		ls.New(),
		lsmod.New(),
		mkdir.New(),
		mknod.New(),
		mount.New(),
		ping.New(),
		ps.New(),
		pwd.New(),
		reboot.New(),
		redisd.New(),
		reload.New(),
		restart.New(),
		rm.New(),
		rmmod.New(),
		show_commands.New(),
		show_packages.New(),
		slashinit.New(),
		sleep.New(),
		source.New(),
		start.New(),
		stop.New(),
		stty.New(),
		subscribe.New(),
		sync.New(),
		umount.New(),
		uninstall.New(),
		wget.New(),
	)
	return g
}
