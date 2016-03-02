// This file is part of go-dicom.
//
// Copyright (C) 2016  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

import (
	// "bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/davidgamba/go-dicom/pdu"
	"github.com/davidgamba/go-dicom/sopclass"
	"github.com/davidgamba/go-dicom/syntax/ts"
	"github.com/davidgamba/go-getoptions" // As getoptions
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

type dicomqr struct {
	CalledAE  [16]byte
	CallingAE [16]byte
	Host      string
	Port      int
	Conn      net.Conn
	ar        pdu.AAssociateRequest
	rr        pdu.AReleaseRequest
}

func (qr *dicomqr) Dial() error {
	address := qr.Host + ":" + strconv.Itoa(qr.Port)
	log.Printf("Connecting to: %s", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	log.Printf("Connecting successful")
	qr.Conn = conn
	return nil
}

func (qr *dicomqr) Init() {
	qr.ar = pdu.AAssociateRequest{
		PDUType:   1,
		CalledAE:  qr.CalledAE,
		CallingAE: qr.CallingAE,
		Content:   []byte{},
	}
	putIntToByteSize2(&qr.ar.ProtocolVersion, 1)

	qr.rr = pdu.AReleaseRequest{
		PDUType: 5,
	}
	qr.rr.Len()
}

func putIntToByteSize2(b *[2]byte, v int) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func (qr *dicomqr) AR() (int, error) {
	if len(qr.ar.Content) == 0 {
		return 0, fmt.Errorf("AR has no content")
	}
	qr.ar.Len()
	b := qr.ar.ToBytes()
	printBytes(b)
	i, err := qr.Conn.Write(b)
	return i, err
}

func (qr *dicomqr) RR() (int, error) {
	b := qr.rr.ToBytes()
	printBytes(b)
	i, err := qr.Conn.Write(b)
	return i, err
}

func (qr *dicomqr) ARAdd(b []byte) {
	qr.ar.Content = append(qr.ar.Content, b...)
}

func padRight(str, pad string, lenght int) string {
	for {
		str += pad
		if len(str) > lenght {
			return str[0:lenght]
		}
	}
}

// AppContextName = "1.2.840.10008.3.1.1.1"
const AppContextName = "1.2.840.10008.3.1.1.1"

// AppContextItem returns a byte slice with app context item.
func AppContextItem() []byte {
	return stringItem([]byte{0x10}, AppContextName, "AppContextItem")
}

// PressContextItem returns a byte slice with press context item.
func PressContextItem(items ...[]byte) []byte {
	b := []byte{32}                               // itemType
	b = append(b, make([]byte, 1)...)             // Reserved
	payload := []byte{1}                          // contextID
	payload = append(payload, make([]byte, 1)...) // Reserved
	payload = append(payload, make([]byte, 1)...) // Result
	payload = append(payload, make([]byte, 1)...) // Reserved
	for _, i := range items {
		payload = append(payload, i...)
	}
	b = append(b, getBytesLenght(2, payload)...) // itemLenght
	b = append(b, payload...)
	fmt.Printf("PressContextItem:\n")
	printBytes(b)
	return b
}

// AbstractSyntaxItem returns a byte slice with abstract syntax item.
func AbstractSyntaxItem() []byte {
	return stringItem([]byte{0x30}, sopclass.VerificationSOPClass, "AbstractSyntaxItem")
}

// TrasnferSyntaxItem returns a byte slice with transfer syntax item.
func TrasnferSyntaxItem(tsi string) []byte {
	return stringItem([]byte{0x40}, tsi, "TrasnferSyntaxItem")
}

func stringItem(t []byte, content, name string) []byte {
	b := t
	b = append(b, make([]byte, 1)...) // Reserved
	payload := []byte(content)
	b = append(b, getStringLenght(2, content)...) // itemLenght
	b = append(b, payload...)
	fmt.Printf("%s:\n", name)
	printBytes(b)
	return b
}

// UserInfoItem returns a byte slice with user info item.
func UserInfoItem(items ...[]byte) []byte {
	b := []byte{0x50}                 // itemType
	b = append(b, make([]byte, 1)...) // Reserved
	payload := []byte{}
	for _, i := range items {
		payload = append(payload, i...)
	}
	b = append(b, getBytesLenght(2, payload)...) // itemLenght
	b = append(b, payload...)
	fmt.Printf("UserInfoItem:\n")
	printBytes(b)
	return b
}

// MaximunLenghtItem returns a byte slice with maximun lenght item.
func MaximunLenghtItem(lenght uint32) []byte {
	b := []byte{0x51}                 // itemType
	b = append(b, make([]byte, 1)...) // Reserved
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, lenght)
	b = append(b, getBytesLenght(2, payload)...) // itemLenght
	b = append(b, payload...)
	fmt.Printf("MaximunLenghtItem:\n")
	printBytes(b)
	return b
}

// ImplementationClassUID = "1.2.40.0.13.1.1"
const ImplementationClassUID = "1.2.40.0.13.1.1"

// ImplementationVersion = "go-dicom-0.1.0"
const ImplementationVersion = "go-dicom-0.1.0"

// ImplementationUIDItem returns a byte slice
func ImplementationUIDItem() []byte {
	b := []byte{82}                   // itemType
	b = append(b, make([]byte, 1)...) // Reserved
	payload := []byte(ImplementationClassUID)
	b = append(b, getStringLenght(2, ImplementationClassUID)...) // itemLenght
	b = append(b, payload...)
	fmt.Printf("ImplementationUID:\n")
	printBytes(b)
	return b
}

// ImplementationVersionItem returns a byte slice
func ImplementationVersionItem() []byte {
	b := []byte{85}                   // itemType
	b = append(b, make([]byte, 1)...) // Reserved
	payload := []byte(ImplementationVersion)
	b = append(b, getStringLenght(2, ImplementationVersion)...) // itemLenght
	b = append(b, payload...)
	fmt.Printf("ImplementationVersion:\n")
	printBytes(b)
	return b
}

func getStringLenght(size int, content string) []byte {
	b := make([]byte, size)
	binary.BigEndian.PutUint16(b, uint16(len(content)))
	return b
}

func getBytesLenght(size int, content []byte) []byte {
	return intToBytes(size, len(content))
}

func intToBytes(size, i int) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, int64(i))
	if err != nil {
		panic(err)
	}
	b := buf.Bytes()
	return b[len(b)-size:]
}

func printBytes(b []byte) {
	l := len(b)
	var s string
	for i := 0; i < l; i++ {
		s += stripCtlFromUTF8(string(b[i]))
		if i != 0 && i%8 == 0 {
			if i%16 == 0 {
				fmt.Printf(" - %s\n", s)
				s = ""
			} else {
				fmt.Printf(" - ")
			}
		}
		fmt.Printf("%2x ", b[i])
		if i == l-1 {
			if 15-i%16 > 7 {
				fmt.Printf(" - ")
			}
			for j := 0; j < 15-i%16; j++ {
				// fmt.Printf("   ")
				fmt.Printf("   ")
			}
			fmt.Printf(" - %s\n", s)
			s = ""
		}
	}
	fmt.Printf("\n")
}

// http://rosettacode.org/wiki/Strip_control_codes_and_extended_characters_from_a_string#Go
// two UTF-8 functions identical except for operator comparing c to 127
func stripCtlFromUTF8(str string) string {
	return strings.Map(func(r rune) rune {
		if r >= 32 && r != 127 {
			return r
		}
		return '.'
	}, str)
}

func byte16PutString(s string) [16]byte {
	var a [16]byte
	if len(s) > 16 {
		copy(a[:], s)
	} else {
		copy(a[16-len(s):], s)
	}
	return a
}

func (qr *dicomqr) HandleAccept() error {
	// Ignore
	tbuf := make([]byte, 1)
	_, err := qr.Conn.Read(tbuf)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	// Read PDULenght
	tbuf = make([]byte, 4)
	_, err = qr.Conn.Read(tbuf)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	var size uint32
	buf := bytes.NewBuffer(tbuf)
	binary.Read(buf, binary.BigEndian, &size)
	fmt.Println(size)

	// Get PDU
	tbuf = make([]byte, size)
	_, err = qr.Conn.Read(tbuf)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	printBytes(tbuf)
	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	var host, ae string
	var port int
	opt := getoptions.GetOptions()
	opt.StringVar(&host, "host", "localhost")
	opt.IntVar(&port, "port", 11112)
	opt.StringVar(&ae, "ae", "PACSAE")
	_, err := opt.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	qr := dicomqr{
		CalledAE:  byte16PutString(ae),
		CallingAE: byte16PutString("go-dicom"),
		Host:      host,
		Port:      port,
	}

	qr.Init()
	fmt.Printf("%v\n", qr.ar)

	qr.ARAdd(AppContextItem())

	qr.ARAdd(PressContextItem(
		AbstractSyntaxItem(),
		TrasnferSyntaxItem(ts.ImplicitVRLittleEndian),
		TrasnferSyntaxItem(ts.ExplicitVRLittleEndian),
		TrasnferSyntaxItem(ts.ExplicitVRBigEndian),
	))
	qr.ARAdd(UserInfoItem(
		MaximunLenghtItem(32768),
		ImplementationUIDItem(),
		ImplementationVersionItem(),
	))

	err = qr.Dial()
	if err != nil {
		log.Fatal(err)
	}
	defer qr.Conn.Close()

	i, err := qr.AR()
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	log.Printf("Payload sent: %d bytes", i)

	tbuf := make([]byte, 1)
	_, err = qr.Conn.Read(tbuf)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	switch tbuf[0] {
	case 0x2:
		fmt.Println("A-ASSOCIATE accept")
		qr.HandleAccept()
	default:
		printBytes(tbuf)
		fmt.Println("Unknown")
		log.Fatal(err)
	}

	i, err = qr.RR()
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	log.Printf("Payload sent: %d bytes", i)

	tbuf = make([]byte, 1)
	_, err = qr.Conn.Read(tbuf)
	if err != nil {
		log.Fatal("Error reading", err)
	}
	switch tbuf[0] {
	case 0x6:
		fmt.Println("A-RELEASE response")
		qr.Conn.Close()
	default:
		printBytes(tbuf)
		fmt.Println("Unknown")
		log.Fatal(err)
	}

	qr.Conn.Close()
}
