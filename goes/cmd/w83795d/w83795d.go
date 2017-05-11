// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package w83795d provides access to the H/W Monitor chip

package w83795d

import (
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
)

const (
	Name    = "w83795d"
	Apropos = "w83795 hardware monitoring daemon, publishes to redis"
	Usage   = "w83795d"
)

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

var (
	Init = func() {}
	once sync.Once

	first int

	Vdev I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)
)

type cmd struct {
	Info
}

type Info struct {
	mutex sync.Mutex
	rpc   *sockfile.RpcServer
	pub   *publisher.Publisher
	stop  chan struct{}
	last  map[string]uint16
	lasts map[string]string
}

func New() *cmd { return new(cmd) }

func (*cmd) Apropos() lang.Alt { return apropos }
func (*cmd) Kind() goes.Kind   { return goes.Daemon }
func (*cmd) String() string    { return Name }
func (*cmd) Usage() string     { return Usage }

func (cmd *cmd) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t
	var err error
	first = 1

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]uint16)
	cmd.lasts = make(map[string]string)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if cmd.rpc, err = sockfile.NewRpcServer(Name); err != nil {
		return err
	}

	rpc.Register(&cmd.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", Name, "Info")
		if err != nil {
			return err
		}
	}

	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if err = cmd.update(); err != nil {
				close(cmd.stop)
				return err
			}
		}
	}
	return nil
}

func (cmd *cmd) Close() error {
	close(cmd.stop)
	return nil
}

func (cmd *cmd) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}
	if err := writeRegs(); err != nil {
		return err
	}

	if first == 1 {
		Vdev.FanInit()
		first = 0
	}

	for k, i := range VpageByKey {
		if strings.Contains(k, "rpm") {
			v, err := Vdev.FanCount(i)
			if err != nil {
				return err
			}
			if v != cmd.last[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.last[k] = v
			}
		}
		if strings.Contains(k, "fan_tray.speed") {
			v, err := Vdev.GetFanSpeed(i)
			if err != nil {
				return err
			}
			if v != cmd.lasts[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.lasts[k] = v
			}
		}
		if strings.Contains(k, "fan_tray.duty") {
			v, err := Vdev.GetFanSpeed(i)
			if err != nil {
				return err
			}
			if v != cmd.lasts[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.lasts[k] = v
			}
		}
		if strings.Contains(k, "hwmon.front.temp.units.C") {
			v, err := Vdev.FrontTemp()
			if err != nil {
				return err
			}
			if v != cmd.lasts[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.lasts[k] = v
			}
		}
		if strings.Contains(k, "hwmon.rear.temp.units.C") {
			v, err := Vdev.RearTemp()
			if err != nil {
				return err
			}
			if v != cmd.lasts[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.lasts[k] = v
			}
		}
	}
	return nil
}

const (
	fanPoles    = 4
	tempCtrl2   = 0x5f
	high        = 0xff
	med         = 0x80
	low         = 0x50
	maxFanTrays = 4
)

func fanSpeed(countHi uint8, countLo uint8) uint16 {
	d := ((uint16(countHi) << 4) + (uint16(countLo & 0xf))) * (uint16(fanPoles / 4))
	speed := 1.35E06 / float64(d)
	return uint16(speed)
}

func (h *I2cDev) FrontTemp() (string, error) {
	r := getRegsBank0()
	r.BankSelect.set(h, 0x80)
	r.FrontTemp.get(h)
	r.FractionLSB.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint8(s[3].D[0])
	u := uint8(s[5].D[0])
	v := float64(t) + ((float64(u >> 7)) * 0.25)
	strconv.FormatFloat(v, 'f', 3, 64)
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) RearTemp() (string, error) {
	r := getRegsBank0()
	r.BankSelect.set(h, 0x80)
	r.RearTemp.get(h)
	r.FractionLSB.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint8(s[3].D[0])
	u := uint8(s[5].D[0])
	v := float64(t) + ((float64(u >> 7)) * 0.25)
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) FanCount(i uint8) (uint16, error) {
	var rpm uint16

	if i > 14 {
		panic("FanCount subscript out of range\n")
	}
	i--

	n := i/2 + 1
	w := "fan_tray." + strconv.Itoa(int(n)) + ".status"
	p, _ := redis.Hget(redis.DefaultHash, w)

	//set fan speed to max and return 0 rpm if fan tray is not present or failed
	if strings.Contains(p, "not installed") {
		rpm = uint16(0)
	} else {
		//remap physical to logical, 0:7 -> 7:0
		i = i + 7 - (2 * i)
		r := getRegsBank0()
		r.BankSelect.set(h, 0x80)
		r.FanCount[i].get(h)
		r.FractionLSB.get(h)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return 0, err
		}
		t := uint8(s[3].D[0])
		u := uint8(s[5].D[0])
		rpm = fanSpeed(t, u)
	}
	return rpm, nil
}

func (h *I2cDev) FanInit() error {

	//reset hwm to default values
	r0 := getRegsBank0()
	r0.BankSelect.set(h, 0x80)
	r0.Configuration.set(h, 0x9c)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
	//set fan speed output to PWM mode
	r2.FanOutputModeControl.set(h, 0x0)
	//set up clk frequency and dividers
	r2.FanPwmPrescale1.set(h, 0x84)
	r2.FanPwmPrescale2.set(h, 0x84)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	//set default speed to auto
	h.SetFanSpeed("auto")

	//enable temperature monitoring
	r0.BankSelect.set(h, 0x80)
	r0.TempCntl2.set(h, tempCtrl2)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	//temperature monitoring requires a delay before readings are valid
	time.Sleep(500 * time.Millisecond)
	r0.BankSelect.set(h, 0x80)
	r0.Configuration.set(h, 0x1d)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	return nil
}

func (h *I2cDev) SetFanSpeed(w string) error {
	r2 := getRegsBank2()

	//if not all fan trays are ok, only allow high setting
	for j := 1; j <= maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j))+".status")
		if p != "" && !strings.Contains(p, "ok") {
			log.Print("warning: fan failure mode, speed fixed at high")
			w = "high"
			break
		}
	}

	switch w {
	case "auto":
		r2.BankSelect.set(h, 0x82)
		//set thermal cruise
		r2.FanControlModeSelect1.set(h, 0x00)
		r2.FanControlModeSelect2.set(h, 0x00)
		//set step up and down time to 1s
		r2.FanStepUpTime.set(h, 0x0a)
		r2.FanStepDownTime.set(h, 0x0a)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}

		r2.BankSelect.set(h, 0x82)
		//set fan start speed
		r2.FanStartValue1.set(h, 0x30)
		r2.FanStartValue2.set(h, 0x30)
		//set fan stop speed
		r2.FanStopValue1.set(h, 0x30)
		r2.FanStopValue2.set(h, 0x30)
		closeMux(h)
		err = DoI2cRpc()
		if err != nil {
			return err
		}

		r2.BankSelect.set(h, 0x82)
		//set fan stop time to never stop
		r2.FanStopTime1.set(h, 0x0)
		r2.FanStopTime2.set(h, 0x0)
		//set target temps to 50°C
		r2.TargetTemp1.set(h, 0x32)
		r2.TargetTemp2.set(h, 0x32)
		closeMux(h)
		err = DoI2cRpc()
		if err != nil {
			return err
		}

		r2.BankSelect.set(h, 0x82)
		//set critical temp to set 100% fan speed to 65°C
		r2.FanCritTemp1.set(h, 0x41)
		r2.FanCritTemp2.set(h, 0x41)
		//set target temp hysteresis to +/- 5°C
		r2.TempHyster1.set(h, 0x55)
		r2.TempHyster2.set(h, 0x55)
		//enable temp control of fans
		r2.TempToFanMap1.set(h, 0xff)
		r2.TempToFanMap2.set(h, 0xff)
		closeMux(h)
		err = DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to ", w)
	//static speed settings below, set hwm to manual mode, then set static speed
	case "high":
		r2.BankSelect.set(h, 0x82)
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, high)
		r2.FanOutValue2.set(h, high)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to ", w)
	case "med":
		r2.BankSelect.set(h, 0x82)
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, med)
		r2.FanOutValue2.set(h, med)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to ", w)
	case "low":
		r2.BankSelect.set(h, 0x82)
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, low)
		r2.FanOutValue2.set(h, low)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to ", w)
	default:
	}

	return nil
}

func (h *I2cDev) GetFanSpeed(i uint8) (string, error) {
	var speed string

	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
	r2.TempToFanMap1.get(h)
	r2.FanOutValue1.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "error", err
	}
	t := uint8(s[3].D[0])
	m := uint8(s[5].D[0])
	if i == 1 {
		return fmt.Sprintf("0x%x", m), nil
	}

	if t == 0xff {
		speed = "auto"
	} else if m == high {
		speed = "high"
	} else if m == med {
		speed = "med"
	} else if m == low {
		speed = "low"
	} else {
		speed = "invalid " + strconv.Itoa(int(m))
	}
	return speed, nil
}

func writeRegs() error {
	for k, v := range WrRegVal {
		switch WrRegFn[k] {
		case "speed":
			if v == "auto" || v == "high" || v == "med" || v == "low" {
				Vdev.SetFanSpeed(v)
			}
		}
		delete(WrRegVal, k)
	}
	return nil
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	_, p := WrRegFn[args.Field]
	if !p {
		return fmt.Errorf("cannot hset: %s", args.Field)
	}
	_, q := WrRegRng[args.Field]
	if !q {
		err := i.set(args.Field, string(args.Value), false)
		if err == nil {
			*reply = 1
			WrRegVal[args.Field] = string(args.Value)
		}
		return err
	}
	for _, v := range WrRegRng[args.Field] {
		if v == string(args.Value) {
			err := i.set(args.Field, string(args.Value), false)
			if err == nil {
				*reply = 1
				WrRegVal[args.Field] = string(args.Value)
			}
			return err
		}
	}
	return fmt.Errorf("Cannot hset.  Valid values are: %s", WrRegRng[args.Field])
}

func (i *Info) set(key, value string, isReadyEvent bool) error {
	i.pub.Print(key, ": ", value)
	return nil
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}
