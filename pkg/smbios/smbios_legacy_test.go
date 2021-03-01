// Copyright 2016-2021 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package smbios

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/u-root/u-root/pkg/memio"
	"github.com/u-root/u-root/pkg/testutil"
)

var tmpBuf = []byte{0, 0, 0, 0, 0, 0}

func mockMemioRead(base int64, uintn memio.UintN) error {
	dat, ok := uintn.(*memio.ByteSlice)
	if !ok {
		return fmt.Errorf("not supported")
	}
	bufLen := len(tmpBuf)
	for i := int64(0); i < dat.Size(); i++ {
		(*dat)[i] = tmpBuf[(base+i)%int64(bufLen)]
	}
	return nil
}

func TestGetSMBIOSLegacyNotFound(t *testing.T) {
	defer func(old func(base int64, uintn memio.UintN) error) { memioRead = old }(memioRead)
	memioRead = mockMemioRead

	_, _, err := GetSMBIOSBaseLegacy()

	want := "could not find _SM_ or _SM3_ via /dev/mem from 0x000f0000 to 0x00100000"
	if err.Error() != want {
		t.Fatalf("want %s, get '%v'", want, err)
	}
}

func TestGetSMBIOSLegacyMemIoReadError(t *testing.T) {
	defer func(old func(base int64, uintn memio.UintN) error) { memioRead = old }(memioRead)
	memioRead = func(base int64, uintn memio.UintN) error {
		return fmt.Errorf("MEMIOREAD_ERROR")
	}

	_, _, err := GetSMBIOSBaseLegacy()

	want := "MEMIOREAD_ERROR"
	if err.Error() != want {
		t.Fatalf("want %s, get '%v'", want, err)
	}
}

func TestGetSMBIOSLegacySMBIOS(t *testing.T) {
	tmpBuf = []byte{0, '_', 'M', 'S', '_', 0, 0, '_', 'S', 'M', '_', 0, 0, 0, 0, 0}
	defer func(old func(base int64, uintn memio.UintN) error) { memioRead = old }(memioRead)
	memioRead = mockMemioRead
	base, size, err := GetSMBIOSBaseLegacy()
	if err != nil {
		t.Fatal(err)
	}

	var want int64 = 0xf0007

	if base != want {
		t.Errorf("GetSMBIOSLegacy() get 0x%x, want 0x%x", base, want)
	}

	var wantSize int64 = 0x1f

	if size != wantSize {
		t.Errorf("GetSMBIOSLegacy() get size 0x%x, want 0x%x", size, wantSize)
	}
}

func TestGetSMBIOSLegacySMBIOS3(t *testing.T) {
	tmpBuf = []byte{0, '_', 'M', 'S', '_', 0, 0, '_', 'S', 'M', '3', '_', 0, 0, 0, 0, 0}
	defer func(old func(base int64, uintn memio.UintN) error) { memioRead = old }(memioRead)
	memioRead = mockMemioRead
	base, size, err := GetSMBIOSBaseLegacy()
	if err != nil {
		t.Fatal(err)
	}

	var want int64 = 0xf0009

	if base != want {
		t.Errorf("GetSMBIOSLegacy() get base 0x%x, want 0x%x", base, want)
	}

	var wantSize int64 = 0x18

	if size != wantSize {
		t.Errorf("GetSMBIOSLegacy() get size 0x%x, want 0x%x", size, wantSize)
	}
}

func TestGetSMBIOSLegacyQEMU(t *testing.T) {
	if runtime.GOARCH != "amd64" {
		t.Skipf("test not supported on %s", runtime.GOARCH)
	}
	testutil.SkipIfNotRoot(t)
	base, size, err := GetSMBIOSBaseLegacy()
	if err != nil {
		t.Fatal(err)
	}

	if base == 0 {
		t.Errorf("GetSMBIOSLegacy() does not get SMBIOS base")
	}

	if size != smbios2HeaderSize && size != smbios3HeaderSize {
		t.Errorf("GetSMBIOSLegacy() get size 0x%x, want 0x%x or 0x%x ", size, smbios2HeaderSize, smbios3HeaderSize)
	}
}