// Package main is a script that reads a filesystem full of dcm files and
// generates a json report.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	// "strconv"
	"log"
	"strings"

	"github.com/davidgamba/go-dicom/dcmdump/tag"
	"github.com/davidgamba/go-getoptions"
)

var debug bool

func debugf(format string, a ...interface{}) (n int, err error) {
	if debug {
		return fmt.Printf(format, a...)
	}
	return 0, nil
}
func debugln(a ...interface{}) (n int, err error) {
	if debug {
		return fmt.Println(a...)
	}
	return 0, nil
}

type stringSlice []string

func (s stringSlice) contains(a string) bool {
	for _, b := range s {
		if a == b {
			return true
		}
	}
	return false
}

type dicomqr struct {
	Empty [128]byte
	DICM  [4]byte
	Rest  []byte
}

type fh os.File

func readNBytes(f *os.File, size int) ([]byte, error) {
	data := make([]byte, size)
	for {
		data = data[:cap(data)]
		n, err := f.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		data = data[:n]
	}
	return data, nil
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

func tagString(b []byte) string {
	tag := strings.ToUpper(fmt.Sprintf("%02x%02x%02x%02x", b[1], b[0], b[3], b[2]))
	debugf("%s", tag)
	return tag
}

func printBytes(b []byte) {
	if !debug {
		return
	}
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
		"fixed": true,
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

func stringData(bytes []byte, vr string) string {
	if _, ok := VR[vr]["fixed"]; ok && VR[vr]["fixed"].(bool) {
		s := ""
		l := len(bytes)
		n := 0
		vrl := VR[vr]["len"].(int)
		switch vrl {
		case 1:
			for n+1 <= l {
				s += fmt.Sprintf("%d ", bytes[n])
				n++
			}
			return s
		case 2:
			for n+2 <= l {
				e := binary.LittleEndian.Uint16(bytes[n : n+2])
				s += fmt.Sprintf("%d ", e)
				n += 2
			}
			return s
		case 4:
			for n+4 <= l {
				e := binary.LittleEndian.Uint32(bytes[n : n+4])
				s += fmt.Sprintf("%d ", e)
				n += 4
			}
			return s
		default:
			return string(bytes)
		}
	} else {
		if _, ok := VR[vr]["padded"]; ok && VR[vr]["padded"].(bool) {
			l := len(bytes)
			if bytes[l-1] == 0x0 {
				return string(bytes[:l-1])
			}
			return string(bytes)
		}
		return string(bytes)
	}
}

func parseDataElement(bytes []byte, n int, explicit bool) {
	log.Printf("parseDataElement")
	l := len(bytes)
	// Data element
	m := n
	for n <= l && m+4 <= l {
		m += 4
		printBytes(bytes[n:m])
		t := bytes[n:m]
		tagStr := tagString(t)
		log.Printf("n: %d, Tag: %X -> %s\n", n, t, tagStr)
		n = m
		if tagStr == "" {
			log.Printf("%d Empty Tag: %s\n", n, tagStr)
		} else if _, ok := tag.Tag[tagStr]; !ok {
			fmt.Fprintf(os.Stderr, "ERROR: %d Missing tag '%s'\n", n, tagStr)
		} else {
			log.Printf("Tag Name: %s\n", tag.Tag[tagStr]["name"])
		}
		var len int
		var vr string
		if explicit {
			debugf("%d VR\n", n)
			m += 2
			printBytes(bytes[n:m])
			vr = string(bytes[n:m])
			if _, ok := VR[vr]; !ok {
				// if bytes[n] == 0x0 && bytes[n+1] == 0x0 {
				// 	fmt.Fprintf(os.Stderr, "ERROR: Blank VR\n")
				// } else {
				fmt.Fprintf(os.Stderr, "ERROR: %d Missing VR '%s'\n", n, vr)
				printBytes(bytes[n:])
				return
				// }
			}
			n = m
			if vr == "OB" ||
				vr == "OD" ||
				vr == "OF" ||
				vr == "OL" ||
				vr == "OW" ||
				vr == "SQ" ||
				vr == "UC" ||
				vr == "UR" ||
				vr == "UT" ||
				vr == "UN" {
				debugln("Reserved")
				m += 2
				printBytes(bytes[n:m])
				n = m
				debugln("Lenght")
				m += 4
				printBytes(bytes[n:m])
				len32 := binary.LittleEndian.Uint32(bytes[n:m])
				len = int(len32)
				n = m
			} else {
				debugln("Lenght")
				m += 2
				printBytes(bytes[n:m])
				len16 := binary.LittleEndian.Uint16(bytes[n:m])
				len = int(len16)
				n = m
			}
		} else {
			debugln("Lenght")
			m += 4
			printBytes(bytes[n:m])
			len32 := binary.LittleEndian.Uint32(bytes[n:m])
			len = int(len32)
			n = m
		}
		debugf("Lenght: %d\n", len)
		if vr == "SQ" {
			n = parseSQDataElement(bytes, n, explicit)
			m = n
		} else {
			m += len
			if len < 128 {
				if _, ok := tag.Tag[tagStr]; !ok {
					fmt.Printf("(%s) %s %s %s\n", tagStr, vr, "MISSING", stringData(bytes[n:m], vr))
				} else {
					fmt.Printf("(%s) %s %s %s\n", tagStr, vr, tag.Tag[tagStr]["name"], stringData(bytes[n:m], vr))
				}
			} else {
				if _, ok := tag.Tag[tagStr]; !ok {
					fmt.Printf("(%s) %s %s %s\n", tagStr, vr, "MISSING", "...")
				} else {
					fmt.Printf("(%s) %s %s %s\n", tagStr, vr, tag.Tag[tagStr]["name"], "...")
				}
			}
			printBytes(bytes[n:m])
			n = m
		}
	}
	log.Printf("parseDataElement Complete")
}

func parseSQDataElement(bytes []byte, n int, explicit bool) int {
	log.Printf("parseSQDataElement")
	// SQ Data element
	sequenceDelimitationItem := false
	for !sequenceDelimitationItem {
		m := n + 4
		t := bytes[n:m]
		printBytes(bytes[n:m])
		tagStr := tagString(t)
		log.Printf("Tag: %X -> %s\n", t, tagStr)
		if _, ok := tag.Tag[tagStr]; !ok {
			fmt.Fprintf(os.Stderr, "ERROR: Missing tag '%s'\n", tagStr)
			return n
		}
		if tag.Tag[tagStr]["name"] == "SequenceDelimitationItem" {
			sequenceDelimitationItem = true
		}
		n = m
		log.Printf("Tag Name: %s\n", tag.Tag[tagStr]["name"])
		var len int
		debugln("Lenght")
		m += 4
		printBytes(bytes[n:m])
		len32 := binary.LittleEndian.Uint32(bytes[n:m])
		len = int(len32)
		n = m
		if len == 4294967295 {
			debugln("Lenght undefined")
			for {
				// Find FFFEE00D: ItemDelimitationItem
				tag := bytes[n : n+4]
				tagStr := tagString(tag)
				if tagStr == "FFFEE00D" {
					m += 4
					printBytes(bytes[n:m])
					log.Printf("Tag: %X -> %s\n", tag, tagStr)
					n = m
					debugln("Item Delim found")
					m += 4
					printBytes(bytes[n:m])
					n = m
					break
				} else {
					m++
					n = m
				}
			}
		} else {
			debugf("Lenght: %d\n", len)
			debugln("Data")
			m += len
			printBytes(bytes[n:m])
			n = m
		}
	}
	log.Printf("parseSQDataElement Complete")
	return n
}

func synopsis() {
	synopsis := `dcmdump <dcm_file> [--debug]
`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {

	var file string
	opt := getoptions.New()
	opt.Bool("help", false)
	opt.BoolVar(&debug, "debug", false)
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if opt.Called("help") {
		synopsis()
		os.Exit(1)
	}
	if len(remaining) < 1 {
		fmt.Fprintf(os.Stderr, "ERROR: Missing file\n")
		synopsis()
		os.Exit(1)
	}
	file = remaining[0]
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: failed to read file: '%s'\n", err)
		os.Exit(1)
	}

	// Intro
	n := 128
	printBytes(bytes[0:n])
	// DICM
	m := n + 4
	printBytes(bytes[n:m])
	n = m

	explicit := true

	parseDataElement(bytes, n, explicit)
}
