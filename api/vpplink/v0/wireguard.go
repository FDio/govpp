// Copyright (C) 2020 Cisco Systems Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"bytes"
	"fmt"
	"net"
	"reflect"
)

type WireguardTunnel struct {
	Addr       net.IP
	Port       uint16
	SwIfIndex  uint32
	PublicKey  []byte
	PrivateKey []byte
}

func (t *WireguardTunnel) String() string {
	return fmt.Sprintf("[%d] %s:%d", t.SwIfIndex, t.Addr, t.Port)
}

type WireguardPeer struct {
	PublicKey           []byte
	Port                uint16
	PersistentKeepalive int
	TableID             uint32
	Addr                net.IP
	SwIfIndex           uint32
	Index               uint32
	AllowedIps          []net.IPNet
}

func (t *WireguardPeer) allowedIpsMap() map[string]bool {
	m := make(map[string]bool)
	for _, aip := range t.AllowedIps {
		m[aip.String()] = true
	}
	return m
}
func (t *WireguardPeer) Equal(o *WireguardPeer) bool {
	if o == nil {
		return false
	}
	if o.Index != t.Index {
		return false
	}
	if !bytes.Equal(o.PublicKey, t.PublicKey) {
		return false
	}
	if o.Port != t.Port {
		return false
	}
	if o.TableID != t.TableID {
		return false
	}
	if o.SwIfIndex != t.SwIfIndex {
		return false
	}
	if !o.Addr.Equal(t.Addr) {
		return false
	}
	if o.PersistentKeepalive != t.PersistentKeepalive {
		return false
	}
	if len(t.AllowedIps) != len(o.AllowedIps) {
		return false
	}
	/* AllowedIps should be unique */
	m := t.allowedIpsMap()
	for _, aip := range o.AllowedIps {
		if _, found := m[aip.String()]; !found {
			return false
		}
	}
	return true
}
func (t *WireguardPeer) AddAllowedIp(addr net.IPNet) {
	m := t.allowedIpsMap()
	if _, found := m[addr.String()]; !found {
		t.AllowedIps = append(t.AllowedIps, addr)
	}
}
func (t *WireguardPeer) DelAllowedIp(addr net.IPNet) {
	allowedIps := make([]net.IPNet, 0)
	for _, aip := range t.AllowedIps {
		if aip.String() != addr.String() {
			allowedIps = append(allowedIps, aip)
		}
	}
	t.AllowedIps = allowedIps
}
func (t *WireguardPeer) String() string {
	s := fmt.Sprintf("[id=%d swif=%d", t.Index, t.SwIfIndex)
	s += fmt.Sprintf(" addr=%s port=%d", t.Addr, t.Port)
	s += fmt.Sprintf(" pubKey=%s", string(t.PublicKey[:]))
	s += StrableListToString(" allowedIps=", t.AllowedIps)
	if t.TableID != 0 {
		s += fmt.Sprintf(" tbl=%d", t.TableID)
	}
	if t.PersistentKeepalive != 1 {
		s += fmt.Sprintf(" ka=%d", t.PersistentKeepalive)
	}
	s += "]"
	return s
}
func StrableListToString(prefix string, arg interface{}) string {
	value := reflect.ValueOf(arg)
	if value.Len() == 0 {
		return ""
	}
	s := fmt.Sprintf("%s[", prefix)
	for i := 0; i < value.Len(); i++ {
		if i > 0 {
			s += ","
		}
		v := reflect.Indirect(value.Index(i))
		m := v.MethodByName("String")
		if !m.IsValid() {
			m = v.Addr().MethodByName("String")
		}
		ret := m.Call(make([]reflect.Value, 0))[0]
		s += ret.Interface().(string)
	}
	return s + "]"
}
