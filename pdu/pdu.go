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
