// Package pdu contains DICOM PDU structures
package pdu

// AReleaseRequest A-Release request
type AReleaseRequest struct {
	PDUType   byte
	Blank     [1]byte
	PDULenght uint32
	Request   [4]byte
}

// AAssociateRequest A-Associate request
type AAssociateRequest struct {
	PDUType         byte
	Blank           [1]byte
	PDULenght       uint32
	ProtocolVersion uint16
	Blank2          [2]byte
	CalledAE        [16]byte
	CallingAE       [16]byte
	Blank3          [32]byte
	Content         []byte
}
