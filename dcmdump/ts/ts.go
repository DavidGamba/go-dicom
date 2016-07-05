package ts

// TS -
// http://www.dicomlibrary.com/dicom/transfer-syntax/
var TS = map[string]map[string]interface{}{
	"1.2.840.10008.1.2": {
		"name": "Implicit VR Endian: Default Transfer Syntax for DICOM",
	},
	"1.2.840.10008.1.2.1": {
		"name": "Explicit VR Little Endian",
	},
	"1.2.840.10008.1.2.1.99": {
		"name": "Deflated Explicit VR Little Endian",
	},
	"1.2.840.10008.1.2.2": {
		"name": "Explicit VR Big Endian",
	},
	"1.2.840.10008.1.2.4.50": {
		"name": "JPEG Baseline (Process 1): Default Transfer Syntax for Lossy JPEG 8-bit Image Compression",
	},
	"1.2.840.10008.1.2.4.51": {
		"name": "JPEG Baseline (Processes 2 & 4): Default Transfer Syntax for Lossy JPEG 12-bit Image Compression (Process 4 only)",
	},
	"1.2.840.10008.1.2.4.52": {
		"name": "JPEG Extended (Processes 3 & 5)	Retired",
	},
	"1.2.840.10008.1.2.4.53": {
		"name": "JPEG Spectral Selection, Nonhierarchical (Processes 6 & 8)	Retired",
	},
	"1.2.840.10008.1.2.4.54": {
		"name": "JPEG Spectral Selection, Nonhierarchical (Processes 7 & 9)	Retired",
	},
	"1.2.840.10008.1.2.4.55": {
		"name": "JPEG Full Progression, Nonhierarchical (Processes 10 & 12)	Retired",
	},
	"1.2.840.10008.1.2.4.56": {
		"name": "JPEG Full Progression, Nonhierarchical (Processes 11 & 13)	Retired",
	},
	"1.2.840.10008.1.2.4.57": {
		"name": "JPEG Lossless, Nonhierarchical (Processes 14)",
	},
	"1.2.840.10008.1.2.4.58": {
		"name": "JPEG Lossless, Nonhierarchical (Processes 15)	Retired",
	},
	"1.2.840.10008.1.2.4.59": {
		"name": "JPEG Extended, Hierarchical (Processes 16 & 18)	Retired",
	},
	"1.2.840.10008.1.2.4.60": {
		"name": "JPEG Extended, Hierarchical (Processes 17 & 19)	Retired",
	},
	"1.2.840.10008.1.2.4.61": {
		"name": "JPEG Spectral Selection, Hierarchical (Processes 20 & 22)	Retired",
	},
	"1.2.840.10008.1.2.4.62": {
		"name": "JPEG Spectral Selection, Hierarchical (Processes 21 & 23)	Retired",
	},
	"1.2.840.10008.1.2.4.63": {
		"name": "JPEG Full Progression, Hierarchical (Processes 24 & 26)	Retired",
	},
	"1.2.840.10008.1.2.4.64": {
		"name": "JPEG Full Progression, Hierarchical (Processes 25 & 27)	Retired",
	},
	"1.2.840.10008.1.2.4.65": {
		"name": "JPEG Lossless, Nonhierarchical (Process 28)	Retired",
	},
	"1.2.840.10008.1.2.4.66": {
		"name": "JPEG Lossless, Nonhierarchical (Process 29)	Retired",
	},
	"1.2.840.10008.1.2.4.70": {
		"name": "JPEG Lossless, Nonhierarchical, First- Order Prediction (Processes 14 [Selection Value 1]): Default Transfer Syntax for Lossless JPEG Image Compression",
	},
	"1.2.840.10008.1.2.4.80": {
		"name": "JPEG-LS Lossless Image Compression",
	},
	"1.2.840.10008.1.2.4.81": {
		"name": "JPEG-LS Lossy (Near- Lossless) Image Compression",
	},
	"1.2.840.10008.1.2.4.90": {
		"name": "JPEG 2000 Image Compression (Lossless Only)",
	},
	"1.2.840.10008.1.2.4.91": {
		"name": "JPEG 2000 Image Compression",
	},
	"1.2.840.10008.1.2.4.92": {
		"name": "JPEG 2000 Part 2 Multicomponent Image Compression (Lossless Only)",
	},
	"1.2.840.10008.1.2.4.93": {
		"name": "JPEG 2000 Part 2 Multicomponent Image Compression",
	},
	"1.2.840.10008.1.2.4.94": {
		"name": "JPIP Referenced",
	},
	"1.2.840.10008.1.2.4.95": {
		"name": "JPIP Referenced Deflate",
	},
	"1.2.840.10008.1.2.5": {
		"name": "RLE Lossless",
	},
	"1.2.840.10008.1.2.6.1": {
		"name": "RFC 2557 MIME Encapsulation",
	},
	"1.2.840.10008.1.2.4.100": {
		"name": "MPEG2 Main Profile Main Level",
	},
	"1.2.840.10008.1.2.4.102": {
		"name": "MPEG-4 AVC/H.264 High Profile / Level 4.1",
	},
	"1.2.840.10008.1.2.4.103": {
		"name": "MPEG-4 AVC/H.264 BD-compatible High Profile / Level 4.1",
	},
}
