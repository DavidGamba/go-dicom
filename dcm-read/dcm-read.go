// Package main is a script that reads a filesystem full of dcm files and
// generates a json report.
// It uses the dcm2xml tool from the dcm4chee 3 toolkit.
// It will use the dcm2xml tool to generate an xml with the dicom attributes
// and then parse through them.
package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"reflect"

	"github.com/davidgamba/go-getoptions"
)

var debug bool

// <NativeDicomModel><DicomAttribute keyword="SeriesDescription" tag="0008103E" vr="LO"><Value number="1">COR T2 HASTE</Value></DicomAttribute></NativeDicomModel>

// NativeDicomModel -
type NativeDicomModel struct {
	DicomAttributes []DicomAttribute `xml:"DicomAttribute"`
}

// DicomAttribute -
type DicomAttribute struct {
	Keyword     string `xml:"keyword,attr"`
	Tag         string `xml:"tag,attr"`
	VR          string `xml:"vr,attr"`
	PatientName string `xml:"PersonName>Alphabetic>FamilyName"`
	Value       []string
}

type dcmKeyType interface{}

// TagFlatList -
type TagFlatList struct {
	PatientTags
	StudyTags
	SeriesTags
	InstanceTags
}

// PatientTags has patient level DICOM tags
type PatientTags struct {
	PatientName              dcmKeyType `dcm:"00100010"`
	PatientID                dcmKeyType `dcm:"00100020"`
	PatientBirthDate         dcmKeyType `dcm:"00100030"`
	NumberOfRelatedStudies   dcmKeyType `dcm:"00201200"`
	NumberOfRelatedSeries    dcmKeyType `dcm:"00201202"`
	NumberOfRelatedInstances dcmKeyType `dcm:"00201204"`
}

// StudyTags has study level DICOM tags
type StudyTags struct {
	StudyInstanceUID         dcmKeyType `dcm:"0020000D"`
	StudyDate                dcmKeyType `dcm:"00080020"`
	AccessionNumber          dcmKeyType `dcm:"00080050"`
	ModalitiesInStudy        dcmKeyType `dcm:"00080061"`
	ReferringPhysicianName   dcmKeyType `dcm:"00080090"`
	StudyDescription         dcmKeyType `dcm:"00081030"`
	NumberOfRelatedSeries    dcmKeyType `dcm:"00201206"`
	NumberOfRelatedInstances dcmKeyType `dcm:"00201208"`
}

// SeriesTags has series level DICOM tags
type SeriesTags struct {
	SeriesInstanceUID        dcmKeyType `dcm:"0020000E"`
	SeriesNumber             dcmKeyType `dcm:"00200011"`
	Modality                 dcmKeyType `dcm:"00080060"`
	NumberOfRelatedInstances dcmKeyType `dcm:"00201209"`
}

// InstanceTags has image level DICOM tags
type InstanceTags struct {
	SOPInstanceUID dcmKeyType `dcm:"00080018"`
	InstanceNumber dcmKeyType `dcm:"00200013"`
}

// GetStructFields - Given a struct, it returns each struct field.
// Each struct field contains a Name, and optionally a Tag.
func GetStructFields(s interface{}) []reflect.StructField {
	t := reflect.TypeOf(s)
	values := []reflect.StructField{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Struct {
			v := reflect.ValueOf(s)
			values = append(values, GetStructFields(v.Field(i).Interface())...)
		} else {
			values = append(values, t.Field(i))
		}
	}
	return values
}

func readDCMFile(bin, dcmFilepath string) (NativeDicomModel, error) {
	var ndm NativeDicomModel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"dcm2xml")
	command = append(command, dcmFilepath)
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] command: %s\n", err)
		return ndm, err
	}
	debugf("%s\n", out)
	err = xml.Unmarshal(out, &ndm)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] unmarshaling XML: %s\n", err)
		return ndm, err
	}
	return ndm, nil
}

func debugf(format string, a ...interface{}) (n int, err error) {
	if debug {
		return fmt.Printf(format, a)
	}
	return 0, nil
}
func debugln(a ...interface{}) (n int, err error) {
	if debug {
		return fmt.Println(a)
	}
	return 0, nil
}

func synopsis() {
	synopsis := `dcm-read <dcm_file_path>
  --lib <dcm4chee-location>
  [--debug]
`
	fmt.Fprintln(os.Stderr, synopsis)
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

func main() {
	var lib string
	opt := getoptions.New()
	opt.StringVar(&lib, "lib", "")
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
	if !opt.Called("lib") {
		fmt.Fprintf(os.Stderr, "[ERROR] missing dcm4chee --lib option\n")
		synopsis()
		os.Exit(1)
	}
	fields := GetStructFields(TagFlatList{})
	tags := []string{}
	for _, f := range fields {
		tags = append(tags, string(f.Tag.Get("dcm")))
	}

	ndm, _ := readDCMFile(lib, remaining[0])
	for _, da := range ndm.DicomAttributes {
		if stringSlice(tags).contains(da.Tag) {
			if len(da.Value) > 0 {
				fmt.Printf("%s -> %s\n", da.Keyword, da.Value[0])
			} else if da.PatientName != "" {
				fmt.Printf("%s -> %v\n", da.Keyword, da.PatientName)
			} else {
				fmt.Printf("%s\n", da.Keyword)
			}
		}
	}
}
