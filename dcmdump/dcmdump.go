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
	vri "github.com/davidgamba/go-dicom/dcmdump/vr"
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

// DataElement -
type DataElement struct {
	N        int
	TagGroup []byte // [2]byte
	TagElem  []byte // [2]byte
	TagStr   string
	VR       []byte // [2]byte
	VRStr    string
	VRLen    int
	Len      uint32
	Data     []byte
	PartOfSQ bool
}

// String -
func (de *DataElement) String() string {
	tn := tag.Tag[de.TagStr]["name"]
	if _, ok := tag.Tag[de.TagStr]; !ok {
		tn = "MISSING"
	}
	if de.Len < 128 {
		return fmt.Sprintf("%04d (%s) %s %d %d %s %s", de.N, de.TagStr, de.VRStr, de.VRLen, de.Len, tn, stringData(de.Data, de.VRStr))
	}
	return fmt.Sprintf("%04d (%s) %s %d %d %s %s", de.N, de.TagStr, de.VRStr, de.VRLen, de.Len, tn, "...")
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

func stringData(bytes []byte, vr string) string {
	if _, ok := vri.VR[vr]["fixed"]; ok && vri.VR[vr]["fixed"].(bool) {
		s := ""
		l := len(bytes)
		n := 0
		vrl := vri.VR[vr]["len"].(int)
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
		if _, ok := vri.VR[vr]["padded"]; ok && vri.VR[vr]["padded"].(bool) {
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
		undefinedLen := false
		de := DataElement{N: n}
		m += 4
		t := bytes[n:m]
		de.TagGroup = bytes[n : n+2]
		de.TagElem = bytes[n+2 : n+4]
		de.TagStr = tagString(t)
		// TODO: Clean up tagString
		tagStr := tagString(t)
		log.Printf("n: %d, Tag: %X -> %s\n", n, t, tagStr)
		printBytes(bytes[n:m])
		n = m
		if tagStr == "" {
			log.Printf("%d Empty Tag: %s\n", n, tagStr)
		} else if _, ok := tag.Tag[tagStr]; !ok {
			fmt.Fprintf(os.Stderr, "ERROR: %d Missing tag '%s'\n", n, tagStr)
		} else {
			log.Printf("Tag Name: %s\n", tag.Tag[tagStr]["name"])
		}
		var len uint32
		var vr string
		if explicit {
			debugf("%d VR\n", n)
			m += 2
			printBytes(bytes[n:m])
			de.VR = bytes[n:m]
			de.VRStr = string(bytes[n:m])
			vr = string(bytes[n:m])
			if _, ok := vri.VR[vr]; !ok {
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
				len = binary.LittleEndian.Uint32(bytes[n:m])
				n = m
			} else {
				debugln("Lenght")
				m += 2
				printBytes(bytes[n:m])
				len16 := binary.LittleEndian.Uint16(bytes[n:m])
				len = uint32(len16)
				n = m
			}
		} else {
			debugln("Lenght")
			m += 4
			printBytes(bytes[n:m])
			len = binary.LittleEndian.Uint32(bytes[n:m])
			n = m
		}
		if len == 0xFFFFFFFF {
			undefinedLen = true
			for {
				// Find FFFEE0DD: SequenceDelimitationItem
				endTag := bytes[m : m+4]
				endTagStr := tagString(endTag)
				if endTagStr == "FFFEE00D" || endTagStr == "FFFEE0DD" {
					log.Printf("found SequenceDelimitationItem")
					len = uint32(m - n)
					m = n
					break
				} else {
					m++
					if m >= l {
						fmt.Fprintf(os.Stderr, "ERROR: Couldn't find SequenceDelimitationItem\n")
						printBytes(bytes[n:l])
						return
					}
				}
			}
		}
		de.Len = len
		debugf("Lenght: %d\n", len)
		m += int(len)
		printBytes(bytes[n:m])
		if vr == "SQ" {
			de.Data = []byte{}
			// n = parseSQDataElement(bytes[n:m], n, explicit)
		} else {
			de.Data = bytes[n:m]
			fmt.Println(de.String())
		}
		if undefinedLen {
			m += 8
		}
		n = m
	}
	log.Printf("parseDataElement Complete")
}

func parseSQDataElement(bytes []byte, n int, explicit bool) int {
	log.Printf("parseSQDataElement")
	// SQ Data element
	sequenceDelimitationItem := false
	for !sequenceDelimitationItem {
		m := n + 4
		printBytes(bytes[n:m])
		t := bytes[n:m]
		tagStr := tagString(t)
		log.Printf("n: %d, Tag: %X -> %s\n", n, t, tagStr)
		if _, ok := tag.Tag[tagStr]; !ok {
			fmt.Fprintf(os.Stderr, "ERROR: %d Missing tag '%s'\n", n, tagStr)
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
				endTag := bytes[n : n+4]
				endTagStr := tagString(endTag)
				if endTagStr == "FFFEE00D" {
					m += 4
					printBytes(bytes[n:m])
					log.Printf("Tag: %X -> %s\n", endTag, endTagStr)
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
			if len < 128 {
				if _, ok := tag.Tag[tagStr]; !ok {
					// TODO: Handle implicit VR
					fmt.Printf("    (%s) %s %s\n", tagStr, "MISSING", stringData(bytes[n:m], "SQ"))
				} else {
					fmt.Printf("    (%s) %s %s\n", tagStr, tag.Tag[tagStr]["name"], stringData(bytes[n:m], "SQ"))
				}
			} else {
				if _, ok := tag.Tag[tagStr]; !ok {
					fmt.Printf("    (%s) %s %s\n", tagStr, "MISSING", "...")
				} else {
					fmt.Printf("    (%s) %s %s\n", tagStr, tag.Tag[tagStr]["name"], "...")
				}
			}
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
