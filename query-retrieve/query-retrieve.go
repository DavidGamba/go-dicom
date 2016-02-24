// Package main provides a tool to query-retrieve initially using dcm4che tools
// but eventually using this library.  The biggest downside of the current
// approach, besides using an external library, is that it creates several
// associations instead of reusing one.
package main

import (
	"fmt"
	"github.com/davidgamba/go-getoptions" // as getoptions
	"gopkg.in/xmlpath.v2"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

var debug bool

type patientLevel struct {
	PatientName              string // 	(0010,0010)
	PatientID                string //	(0010,0020)
	NumberOfRelatedStudies   string //	(0020,1200)
	NumberOfRelatedSeries    string //	(0020,1202)
	NumberOfRelatedInstances string //	(0020,1204)
}

type studyLevel struct {
	StudyInstanceUID         string //	(0020,000D)
	AccessionNumber          string //	(0008,0050)
	ModalitiesInStudy        string //	(0008,0061)
	NumberOfRelatedSeries    string //	(0020,1206)
	NumberOfRelatedInstances string //	(0020,1208)
}

type seriesLevel struct {
	SeriesInstanceUID        string //	(0020,000E)
	SeriesNumber             string //	(0020,0011)
	Modality                 string //	(0008,0060)
	NumberOfRelatedInstances string //	(0020,1209)
}

type instanceLevel struct {
	SOPInstanceUID string //	(0000,0000)
	Modality       string //	(0000,0000)
	InstanceNumber string //	(0000,0000)
}

type byPatientName []patientLevel

func (a byPatientName) Len() int           { return len(a) }
func (a byPatientName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPatientName) Less(i, j int) bool { return a[i].PatientName < a[j].PatientName }

type byStudyInstanceUID []studyLevel

func (a byStudyInstanceUID) Len() int           { return len(a) }
func (a byStudyInstanceUID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStudyInstanceUID) Less(i, j int) bool { return a[i].StudyInstanceUID < a[j].StudyInstanceUID }

type bySeriesNumber []seriesLevel

func (a bySeriesNumber) Len() int      { return len(a) }
func (a bySeriesNumber) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a bySeriesNumber) Less(i, j int) bool {
	in, err := strconv.Atoi(a[i].SeriesNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] SeriesNumber is not numeral: %s", a[i].SeriesNumber)
		return false
	}
	jn, err := strconv.Atoi(a[j].SeriesNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] SeriesNumber is not numeral: %s", a[j].SeriesNumber)
		return false
	}
	return in < jn
}

type byInstanceNumber []instanceLevel

func (a byInstanceNumber) Len() int      { return len(a) }
func (a byInstanceNumber) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byInstanceNumber) Less(i, j int) bool {
	in, err := strconv.Atoi(a[i].InstanceNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] InstanceNumber is not numeral: %s", a[i].InstanceNumber)
		return false
	}
	jn, err := strconv.Atoi(a[j].InstanceNumber)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] InstanceNumber is not numeral: %s", a[j].InstanceNumber)
		return false
	}
	return in < jn
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

func patientList(bin, pacs, bind, dir string) ([]patientLevel, error) {
	return patientFind(bin, pacs, bind, dir, "*")
}

func patientFind(bin, pacs, bind, dir, patient string) ([]patientLevel, error) {
	var pl []patientLevel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"findscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	command = append(command, "-m", "PatientName="+patient)
	command = append(command, "-r", "00100020") // PatientID
	command = append(command, "-r", "00201200") // NumberOfPatientRelatedStudies
	command = append(command, "-r", "00201202") // NumberOfPatientRelatedSeries
	command = append(command, "-r", "00201204") // NumberOfPatientRelatedInstances
	command = append(command, "-L", "PATIENT")
	command = append(command, "-X", "--out-dir", dir, "--out-cat")
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] command: %s\n", err)
		return pl, err
	}
	debugf("%s\n", out)
	f, err := os.Open(dir + string(os.PathSeparator) + "001.dcm")
	if err != nil {
		return pl, err
	}
	root, err := xmlpath.Parse(f)
	if err != nil {
		return pl, err
	}
	path := xmlpath.MustCompile("/NativeDicomModel")
	pnPath := xmlpath.MustCompile("DicomAttribute[@keyword='PatientName']/PersonName[@number='1']/Alphabetic/FamilyName")
	pIDPath := xmlpath.MustCompile("DicomAttribute[@keyword='PatientID']/Value")
	pNSPath := xmlpath.MustCompile("DicomAttribute[@keyword='NumberOfPatientRelatedStudies']/Value")
	pNSerPath := xmlpath.MustCompile("DicomAttribute[@keyword='NumberOfPatientRelatedSeries']/Value")
	pNInsPath := xmlpath.MustCompile("DicomAttribute[@keyword='NumberOfPatientRelatedInstances']/Value")
	iter := path.Iter(root)
	for iter.Next() {
		pn, _ := pnPath.String(iter.Node())
		pID, _ := pIDPath.String(iter.Node())
		pNS, _ := pNSPath.String(iter.Node())
		pNSer, _ := pNSerPath.String(iter.Node())
		pNIns, _ := pNInsPath.String(iter.Node())
		pl = append(pl, patientLevel{PatientName: pn, PatientID: pID, NumberOfRelatedStudies: pNS, NumberOfRelatedSeries: pNSer, NumberOfRelatedInstances: pNIns})
		debugf("%v\n", pl)
	}
	sort.Stable(byPatientName(pl))
	return pl, nil
}

func studyList(bin, pacs, bind, dir, patient string) ([]studyLevel, error) {
	var sl []studyLevel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"findscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	command = append(command, "-m", "PatientName="+patient)
	command = append(command, "-r", "StudyInstanceUID")
	command = append(command, "-r", "00080050") // AccessionNumber
	command = append(command, "-r", "00080061") // ModalitiesInStudy
	command = append(command, "-r", "00201206") // NumberOfStudyRelatedSeries
	command = append(command, "-r", "00201208") // NumberOfStudyRelatedInstances
	command = append(command, "-L", "STUDY")
	command = append(command, "-X", "--out-dir", dir, "--out-cat")
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] command: %s\n", err)
		return sl, err
	}
	debugf("%s\n", out)
	f, err := os.Open(dir + string(os.PathSeparator) + "001.dcm")
	if err != nil {
		return sl, err
	}
	root, err := xmlpath.Parse(f)
	if err != nil {
		return sl, err
	}
	path := xmlpath.MustCompile("/NativeDicomModel")
	suidPath := xmlpath.MustCompile("DicomAttribute[@keyword='StudyInstanceUID']/Value")
	sanPath := xmlpath.MustCompile("DicomAttribute[@keyword='AccessionNumber']/Value")
	smodPath := xmlpath.MustCompile("DicomAttribute[@keyword='ModalitiesInStudy']/Value")
	sNSerPath := xmlpath.MustCompile("DicomAttribute[@keyword='NumberOfStudyRelatedSeries']/Value")
	sNInsPath := xmlpath.MustCompile("DicomAttribute[@keyword='NumberOfStudyRelatedInstances']/Value")
	iter := path.Iter(root)
	for iter.Next() {
		suid, _ := suidPath.String(iter.Node())
		san, _ := sanPath.String(iter.Node())
		smod, _ := smodPath.String(iter.Node())
		sNSer, _ := sNSerPath.String(iter.Node())
		sNIns, _ := sNInsPath.String(iter.Node())
		sl = append(sl,
			studyLevel{StudyInstanceUID: suid,
				AccessionNumber:          san,
				ModalitiesInStudy:        smod,
				NumberOfRelatedSeries:    sNSer,
				NumberOfRelatedInstances: sNIns})
		debugf("%v\n", sl)
	}
	sort.Stable(byStudyInstanceUID(sl))
	return sl, nil
}

func seriesList(bin, pacs, bind, dir, patient, studyUID string) ([]seriesLevel, error) {
	var sl []seriesLevel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"findscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	command = append(command, "-m", "PatientName="+patient)
	command = append(command, "-m", "StudyInstanceUID="+studyUID)
	command = append(command, "-r", "SeriesInstanceUID")
	command = append(command, "-r", "00200011") // SeriesNumber
	command = append(command, "-r", "00080060") // Modality
	command = append(command, "-r", "00201209") // NumberOfSeriesRelatedInstances
	command = append(command, "-L", "SERIES")
	command = append(command, "-X", "--out-dir", dir, "--out-cat")
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] command: %s\n", err)
		return sl, err
	}
	debugf("%s\n", out)
	f, err := os.Open(dir + string(os.PathSeparator) + "001.dcm")
	if err != nil {
		return sl, err
	}
	root, err := xmlpath.Parse(f)
	if err != nil {
		return sl, err
	}
	path := xmlpath.MustCompile("/NativeDicomModel")
	suidPath := xmlpath.MustCompile("DicomAttribute[@keyword='SeriesInstanceUID']/Value")
	snPath := xmlpath.MustCompile("DicomAttribute[@keyword='SeriesNumber']/Value")
	smodPath := xmlpath.MustCompile("DicomAttribute[@keyword='Modality']/Value")
	sNInsPath := xmlpath.MustCompile("DicomAttribute[@keyword='NumberOfSeriesRelatedInstances']/Value")
	iter := path.Iter(root)
	for iter.Next() {
		suid, _ := suidPath.String(iter.Node())
		sn, _ := snPath.String(iter.Node())
		smod, _ := smodPath.String(iter.Node())
		sNIns, _ := sNInsPath.String(iter.Node())
		sl = append(sl,
			seriesLevel{SeriesInstanceUID: suid,
				SeriesNumber:             sn,
				Modality:                 smod,
				NumberOfRelatedInstances: sNIns})
		debugf("%v\n", sl)
	}
	sort.Stable(bySeriesNumber(sl))
	return sl, nil
}

func sopList(bin, pacs, bind, dir, patient, studyUID, seriesUID string) ([]instanceLevel, error) {
	var sl []instanceLevel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"findscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	command = append(command, "-m", "PatientName="+patient)
	command = append(command, "-m", "StudyInstanceUID="+studyUID)
	command = append(command, "-m", "SeriesInstanceUID="+seriesUID)
	command = append(command, "-r", "SOPInstanceUID")
	command = append(command, "-r", "Modality")
	command = append(command, "-r", "InstanceNumber")
	command = append(command, "-L", "IMAGE")
	command = append(command, "-X", "--out-dir", dir, "--out-cat")
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] command: %s\n", err)
		return sl, err
	}
	debugf("%s\n", out)
	f, err := os.Open(dir + string(os.PathSeparator) + "001.dcm")
	if err != nil {
		return sl, err
	}
	root, err := xmlpath.Parse(f)
	if err != nil {
		return sl, err
	}
	path := xmlpath.MustCompile("/NativeDicomModel")
	sopPath := xmlpath.MustCompile("DicomAttribute[@keyword='SOPInstanceUID']/Value")
	modalityPath := xmlpath.MustCompile("DicomAttribute[@keyword='Modality']/Value")
	numberPath := xmlpath.MustCompile("DicomAttribute[@keyword='InstanceNumber']/Value")
	iter := path.Iter(root)
	for iter.Next() {
		sop, _ := sopPath.String(iter.Node())
		modality, _ := modalityPath.String(iter.Node())
		number, _ := numberPath.String(iter.Node())
		sl = append(sl, instanceLevel{SOPInstanceUID: sop, Modality: modality, InstanceNumber: number})
		debugf("%v\n", sl)
	}
	sort.Stable(byInstanceNumber(sl))
	return sl, nil
}

func getSeries(bin, pacs, bind, dir, patient, studyUID, seriesUID string) error {
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"getscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	command = append(command, "-m", "PatientName="+patient)
	command = append(command, "-m", "StudyInstanceUID="+studyUID)
	command = append(command, "-m", "SeriesInstanceUID="+seriesUID)
	command = append(command, "-L", "SERIES")
	command = append(command, "--directory", "data"+string(os.PathSeparator)+
		patient+string(os.PathSeparator)+
		studyUID+string(os.PathSeparator)+
		seriesUID)
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		return err
	}
	for _, s := range strings.Split(string(out), "\n") {
		if strings.Contains(s, "C-GET-RSP [pcid=1, completed=1, failed=0, warning=0, status=0H") {
			return nil
		}
	}
	debugf("%s\n", out)
	return nil
}

func getInstance(bin, pacs, bind, dir, patient, studyUID, seriesUID, sopUID string) error {
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"getscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	command = append(command, "-m", "PatientName="+patient)
	command = append(command, "-m", "StudyInstanceUID="+studyUID)
	command = append(command, "-m", "SeriesInstanceUID="+seriesUID)
	command = append(command, "-m", "SOPInstanceUID="+sopUID)
	command = append(command, "-L", "IMAGE")
	command = append(command, "--directory", "data"+string(os.PathSeparator)+
		patient+string(os.PathSeparator)+
		studyUID+string(os.PathSeparator)+
		seriesUID)
	debugln(command)
	out, err := exec.Command(command[0], command[1:]...).Output()
	if err != nil {
		return err
	}
	for _, s := range strings.Split(string(out), "\n") {
		if strings.Contains(s, "C-GET-RSP [pcid=1, completed=1, failed=0, warning=0, status=0H") {
			return nil
		}
	}
	debugf("%s\n", out)
	return nil
}

func printPatientSOPList(bin, pacs, bind, dir, patient string, level int, get bool) error {
	pl, err := patientFind(bin, pacs, bind, dir, patient)
	if err != nil {
		return err
	}
	for _, p := range pl {
		fmt.Printf("{ PatientName: %s,\n", p.PatientName)
		fmt.Printf("  PatientID: %s,\n", p.PatientID)
		fmt.Printf("  NumberOfPatientRelatedStudies: %s,\n", p.NumberOfRelatedStudies)
		fmt.Printf("  NumberOfPatientRelatedSeries: %s,\n", p.NumberOfRelatedSeries)
		fmt.Printf("  NumberOfPatientRelatedInstances: %s,\n", p.NumberOfRelatedInstances)
		if level >= 1 { // study
			fmt.Printf("  studies: [\n")
			sl, err := studyList(bin, pacs, bind, dir, patient)
			if err != nil {
				return err
			}
			for _, s := range sl {
				fmt.Printf("    { StudyInstanceUID: %s,\n", s.StudyInstanceUID)
				fmt.Printf("      AccessionNumber: %s,\n", s.AccessionNumber)
				fmt.Printf("      ModalitiesInStudy: %s,\n", s.ModalitiesInStudy)
				fmt.Printf("      NumberOfRelatedSeries: %s,\n", s.NumberOfRelatedSeries)
				fmt.Printf("      NumberOfRelatedInstances: %s,\n", s.NumberOfRelatedInstances)
				if level >= 2 { //  series
					fmt.Printf("      series: [\n")
					sel, err := seriesList(bin, pacs, bind, dir, patient, s.StudyInstanceUID)
					if err != nil {
						return err
					}
					for _, se := range sel {
						fmt.Printf("        { SeriesInstanceUID: %s,\n", se.SeriesInstanceUID)
						fmt.Printf("          SeriesNumber: %s,\n", se.SeriesNumber)
						fmt.Printf("          Modality: %s,\n", se.Modality)
						fmt.Printf("          NumberOfRelatedInstances: %s,\n", se.NumberOfRelatedInstances)
						if level == 2 && get {
							getSeries(bin, pacs, dir, bind, patient, s.StudyInstanceUID, se.SeriesInstanceUID)
						}
						if level >= 3 { //  image
							fmt.Printf("          instances: [\n")
							sopl, err := sopList(bin, pacs, bind, dir, patient, s.StudyInstanceUID, se.SeriesInstanceUID)
							if err != nil {
								return err
							}
							for _, sop := range sopl {
								fmt.Printf("            { Modality: %s, SOPInstanceUID: %s,  InstanceNumber: %s}\n", sop.Modality, sop.SOPInstanceUID, sop.InstanceNumber)
							}
							fmt.Printf("          ],\n")
							fmt.Printf("          instanceCount: %d,\n", len(sopl))
						}
						fmt.Printf("        },\n")
					}
					fmt.Printf("      ],\n")
					fmt.Printf("      seriesCount: %d,\n", len(sel))
				}
				fmt.Printf("    },\n")
			}
			fmt.Printf("  ],\n")
			fmt.Printf("  studyCount: %d,\n", len(sl))
		}
		fmt.Printf("}\n")
	}
	return nil
}

func synopsis() {
	synopsis := `query-retrieve <action> [<patient>]
  --lib <dcm4chee-location> -c|--connect <ae@host:port>
  [--bind <callingAE>]
  [--level <1 study, 2 series, 3 image>]
  [--dir <tmp-dir-location>]
  [--debug]

query-retrieve -h # show this help

# actions:
list-all
list-patient <patient>
get-patient <patient>`
	fmt.Fprintln(os.Stderr, synopsis)
}

func main() {
	var help bool
	var pacs, bind, lib, dir string
	var level int
	opt := getoptions.GetOptions()
	opt.BoolVar(&help, "help", false)
	opt.BoolVar(&debug, "debug", false)
	opt.StringVar(&pacs, "connect", "DCM4CHEE@localhost:11112")
	opt.StringVar(&lib, "lib", "")
	opt.StringVar(&bind, "bind", "go-dicom")
	opt.StringVar(&dir, "dir", "tmp")
	opt.IntVar(&level, "level", 3)
	remaining, err := opt.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] getoptions: %s\n", err)
		os.Exit(1)
	}

	if help {
		synopsis()
		os.Exit(1)
	}

	if !opt.Called["connect"] {
		fmt.Fprintf(os.Stderr, "[ERROR] missing --connect option\n")
		synopsis()
		os.Exit(1)
	}
	if !opt.Called["lib"] {
		fmt.Fprintf(os.Stderr, "[ERROR] missing dcm4chee --lib option\n")
		synopsis()
		os.Exit(1)
	}

	if len(remaining) < 1 {
		fmt.Printf("[ERROR] Missing action!\n")
		synopsis()
		os.Exit(1)
	}
	action := remaining[0]

	switch action {
	case "list-all":
		pl, err := patientList(lib, pacs, bind, dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] patientList: %s\n", err)
			os.Exit(1)
		}
		for _, p := range pl {
			err := printPatientSOPList(lib, pacs, bind, dir, p.PatientName, level, false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] printPatientSOPList: %s\n", err)
				os.Exit(1)
			}
		}
	case "list-patient":
		if len(remaining) < 2 {
			fmt.Printf("[ERROR] Missing patient!\n")
			synopsis()
			os.Exit(1)
		}
		patient := remaining[1]
		err := printPatientSOPList(lib, pacs, bind, dir, patient, level, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] printPatientSOPList: %s\n", err)
			os.Exit(1)
		}
	case "get-patient":
		if len(remaining) < 2 {
			fmt.Printf("[ERROR] Missing patient!\n")
			synopsis()
			os.Exit(1)
		}
		patient := remaining[1]
		err := printPatientSOPList(lib, pacs, bind, dir, patient, 2, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] printPatientSOPList: %s\n", err)
			os.Exit(1)
		}
	}
}
