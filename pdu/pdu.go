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

// BigEndian Int to [4]byte
func putIntToByteSize4(b *[4]byte, v uint32) {
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
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
