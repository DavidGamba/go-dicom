package vr

// VR -
// http://dicom.nema.org/medical/dicom/current/output/html/part05.html#table_6.2-1
var VR = map[string]map[string]interface{}{
	"AE": {
		"name":  "Application Entity",
		"len":   16, // maximum?
		"fixed": false,
	},
	"AS": {
		"name":  "Age String",
		"len":   4,
		"fixed": true,
	},
	"AT": {
		"name":  "Attribute Tag",
		"len":   4,
		"fixed": true,
	},
	"CS": {
		"name":  "Code String",
		"len":   16, // maximum?
		"fixed": false,
	},
	"DA": {
		"name":  "Date",
		"len":   8,
		"fixed": false,
	},
	"DS": {
		"name":  "Decimal String",
		"len":   16, // maximum?
		"fixed": false,
	},
	"DT": {
		"name":  "Date Time",
		"len":   26, // maximum?
		"fixed": false,
	},
	"FL": {
		"name":  "Floating Point Single",
		"len":   4,
		"fixed": true,
	},
	"FD": {
		"name":  "Floating Point Double",
		"len":   8,
		"fixed": true,
	},
	"IS": {
		"name":  "Integer String",
		"len":   12, // maximum?
		"fixed": false,
	},
	"LO": {
		"name":  "Long String",
		"len":   64, // maximum?
		"fixed": false,
	},
	"LT": {
		"name":  "Long Text",
		"len":   10240, // maximum?
		"fixed": false,
	},
	"OB": {
		"name":  "Other Byte",
		"len":   1, // see Transfer Syntax definition
		"fixed": true,
	},
	"OD": {
		"name":  "Other Double",
		"len":   8, // see Transfer Syntax definition
		"fixed": true,
	},
	"OF": {
		"name":  "Other Float",
		"len":   4, // see Transfer Syntax definition
		"fixed": true,
	},
	"OL": {
		"name":  "Other Long",
		"len":   4, // see Transfer Syntax definition
		"fixed": true,
	},
	"OW": {
		"name":  "Other Word",
		"len":   2, // see Transfer Syntax definition
		"fixed": true,
	},
	"PN": {
		"name":  "Person Name",
		"len":   64, // maximum per component
		"fixed": false,
	},
	"SH": {
		"name":  "Short String",
		"len":   16, // maximum
		"fixed": false,
	},
	"SL": {
		"name":  "Signed Long",
		"len":   4,
		"fixed": true,
	},
	"SQ": {
		"name":  "Sequence of Items",
		"len":   4,
		"fixed": false,
	},
	"SS": {
		"name":  "Signed Short",
		"len":   2,
		"fixed": true,
	},
	"ST": {
		"name":  "Short Text",
		"len":   1024, // maximum
		"fixed": false,
	},
	"TM": {
		"name":  "Time",
		"len":   14, // maximum
		"fixed": false,
	},
	"UC": {
		"name":  "Unlimited Characters",
		"len":   2 ^ 32 - 2, // maximum
		"fixed": false,
	},
	"UI": {
		"name":   "Unique Identifier (UID)",
		"len":    64, // maximum
		"fixed":  false,
		"padded": true,
	},
	"UL": {
		"name":  "Unsigned Long",
		"len":   4,
		"fixed": true,
	},
	// UN
	"UR": {
		"name":  "Universal Resource Identifier or Universal Resource Locator (URI/URL)",
		"len":   2 ^ 32 - 2,
		"fixed": false,
	},
	"US": {
		"name":  "Unsigned Short",
		"len":   2,
		"fixed": true,
	},
	// UT
}
