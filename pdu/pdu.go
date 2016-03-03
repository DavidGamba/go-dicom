// Package pdu contains DICOM PDU structures
package pdu

// AReleaseRequest A-Release request
type AReleaseRequest struct {
	PDUType   byte
	Blank     [1]byte
	PDULenght [4]byte
	Request   [4]byte
}

// AAssociateRequest A-Associate request
type AAssociateRequest struct {
	PDUType         byte
	Blank           [1]byte
	PDULenght       [4]byte
	ProtocolVersion [2]byte
	Blank2          [2]byte
	CalledAE        [16]byte
	CallingAE       [16]byte
	Blank3          [32]byte
	Content         []byte
}

// PDATATFPDU P-DATA-TF PDU
type PDATATFPDU struct {
	PDUType   byte
	Blank     [1]byte
	PDULenght [4]byte
	Content   []PDVItem
}

// PDVItem Presentation Data Value Item
type PDVItem struct {
	Lenght        [4]byte
	PresContextID byte   // Odd Integers between 1 and 255
	Content       []byte // PDV, DICOM message, Command or Data Set Information
}

// AppContextItem Application Context Item
type AppContextItem struct {
	ItemType       byte
	Blank          [1]byte
	Lenght         [2]byte
	AppContextName []byte // Only One <= 64 bytes
}

// AbstractSyntaxItem Abstract Syntax Item
type AbstractSyntaxItem struct {
	ItemType       byte
	Blank          [1]byte
	Lenght         [2]byte
	AbstractSyntax []byte // Only One in RQ, not present in AC <= 64 bytes
}

// BigEndian Int to [2]byte
func putIntToByteSize2(b *[2]byte, v uint16) {
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

// BigEndian Int to [4]byte
func putIntToByteSize4(b *[4]byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

// Len get the len of AppContextItem
func (e *AppContextItem) Len() {
	l := len(e.AppContextName)
	putIntToByteSize2(&e.Lenght, uint16(l))
}

// Len get the len of AbstractSyntaxItem
func (e *AbstractSyntaxItem) Len() {
	l := len(e.AbstractSyntax)
	putIntToByteSize2(&e.Lenght, uint16(l))
}

// Len get the len of AAssociateRequest
func (e *AAssociateRequest) Len() {
	l := len(e.Content)
	l += 2 + 2 + 16 + 16 + 32
	putIntToByteSize4(&e.PDULenght, uint32(l))
}

// Len get the len of AReleaseRequest
func (e *AReleaseRequest) Len() {
	putIntToByteSize4(&e.PDULenght, uint32(4))
}

// Len get the len of PDATATFPDU
func (e *PDATATFPDU) Len() {
	var l int
	for _, c := range e.Content {
		l += c.Len() + 5
	}
	putIntToByteSize4(&e.PDULenght, uint32(l))
}

// Len get the len of PDVItem
func (e *PDVItem) Len() int {
	l := len(e.Content)
	l++
	putIntToByteSize4(&e.Lenght, uint32(l))
	return l
}

// ToBytes converts AAssociateRequest into []byte
func (e *AAssociateRequest) ToBytes() []byte {
	b := []byte{}
	b = append(b, e.PDUType)
	b = append(b, e.Blank[:]...)
	b = append(b, e.PDULenght[:]...)
	b = append(b, e.ProtocolVersion[:]...)
	b = append(b, e.Blank2[:]...)
	b = append(b, e.CalledAE[:]...)
	b = append(b, e.CallingAE[:]...)
	b = append(b, e.Blank3[:]...)
	b = append(b, e.Content...)
	return b
}

// ToBytes converts AReleaseRequest into []byte
func (e *AReleaseRequest) ToBytes() []byte {
	b := []byte{}
	b = append(b, e.PDUType)
	b = append(b, e.Blank[:]...)
	b = append(b, e.PDULenght[:]...)
	b = append(b, e.Request[:]...)
	return b
}

// ToBytes converts AppContextItem into []byte
func (e *AppContextItem) ToBytes() []byte {
	b := []byte{}
	b = append(b, e.ItemType)
	b = append(b, e.Blank[:]...)
	b = append(b, e.Lenght[:]...)
	b = append(b, e.AppContextName[:]...)
	return b
}

// ToBytes converts AbstractSyntaxItem into []byte
func (e *AbstractSyntaxItem) ToBytes() []byte {
	b := []byte{}
	b = append(b, e.ItemType)
	b = append(b, e.Blank[:]...)
	b = append(b, e.Lenght[:]...)
	b = append(b, e.AbstractSyntax[:]...)
	return b
}

// AppContext returns a byte slice with app context item
func AppContext(name string) []byte {
	e := AppContextItem{
		ItemType:       0x10,
		AppContextName: []byte(name),
	}
	e.Len()
	return e.ToBytes()
}

// AbstractSyntax returns a byte slice with abstract syntax item
func AbstractSyntax(name string) []byte {
	e := AbstractSyntaxItem{
		ItemType:       0x30,
		AbstractSyntax: []byte(name),
	}
	e.Len()
	return e.ToBytes()
}
