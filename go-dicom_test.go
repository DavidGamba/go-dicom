// This file is part of go-dicom.
//
// Copyright (C) 2016  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	"github.com/davidgamba/go-dicom/pdu"
	"github.com/davidgamba/go-dicom/syntax/ts"
	"reflect"
	"testing"
)

func TestIntToBytes(t *testing.T) {
	b := intToBytes(2, 80)
	if !reflect.DeepEqual(b, []byte{0, 0x50}) {
		t.Errorf("Fail: %x", b)
	}
	b = intToBytes(2, 16352)
	if !reflect.DeepEqual(b, []byte{0x3f, 0xe0}) {
		t.Errorf("Fail: %x", b)
	}
}

func TestPutIntToByteSize(t *testing.T) {
	b := [2]byte{}
	putIntToByteSize2(&b, 11)
	if !reflect.DeepEqual(b, [2]byte{0, 0xb}) {
		t.Errorf("Fail: %x", b)
	}
	putIntToByteSize2(&b, 512)
	if !reflect.DeepEqual(b, [2]byte{0x2, 0}) {
		t.Errorf("Fail: %x", b)
	}
	c := [4]byte{}
	pdu.PutIntToByteSize4(&c, 11)
	if !reflect.DeepEqual(c, [4]byte{0, 0, 0, 0xb}) {
		t.Errorf("Fail: %x", c)
	}
	pdu.PutIntToByteSize4(&c, 512)
	if !reflect.DeepEqual(c, [4]byte{0, 0, 0x2, 0}) {
		t.Errorf("Fail: %x", c)
	}
}

func TestTS(t *testing.T) {
	b := TrasnferSyntaxItem(ts.ImplicitVRLittleEndian)
	ts := []byte{0x40, 0, 0, 0x11, 0x31, 0x2e, 0x32, 0x2e, 0x38, 0x34, 0x30, 0x2e, 0x31, 0x30, 0x30, 0x30, 0x38, 0x2e, 0x31, 0x2e, 0x32}
	if !reflect.DeepEqual(b, ts) {
		t.Errorf("Fail: %x", b)
	}
}
