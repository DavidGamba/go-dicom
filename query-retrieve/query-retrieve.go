// Package main provides a tool to query-retrieve initially using dcm4che tools
// but eventually using this library.  The biggest downside of the current
// approach, besides using an external library, is that it creates several
// associations instead of reusing one.
package main

import (
	"fmt"
	"github.com/davidgamba/go-dicom/tag"
	"github.com/davidgamba/go-getoptions" // as getoptions
	"gopkg.in/xmlpath.v2"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

var debug bool
var hideInstances bool

type byPatientName []tag.PatientLevel

func (a byPatientName) Len() int           { return len(a) }
func (a byPatientName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPatientName) Less(i, j int) bool { return a[i].PatientName < a[j].PatientName }

type byStudyInstanceUID []tag.StudyLevel

func (a byStudyInstanceUID) Len() int           { return len(a) }
func (a byStudyInstanceUID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byStudyInstanceUID) Less(i, j int) bool { return a[i].StudyInstanceUID < a[j].StudyInstanceUID }

type bySeriesNumber []tag.SeriesLevel

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

type byInstanceNumber []tag.InstanceLevel

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

func patientLevelFind(bin, pacs, bind, dir string, query ...string) ([]tag.PatientLevel, error) {
	var pl []tag.PatientLevel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"findscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	for _, q := range query {
		command = append(command, "-m", q)
	}
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
		pl = append(pl, tag.PatientLevel{
			PatientName:              pn,
			PatientID:                pID,
			NumberOfRelatedStudies:   pNS,
			NumberOfRelatedSeries:    pNSer,
			NumberOfRelatedInstances: pNIns,
		})
		debugf("%v\n", pl)
	}
	sort.Stable(byPatientName(pl))
	return pl, nil
}

func studyLevelFind(bin, pacs, bind, dir string, studyRoot bool, query ...string) ([]tag.StudyLevel, error) {
	var sl []tag.StudyLevel
	command := []string{}
	command = append(command, bin+string(os.PathSeparator)+"bin"+string(os.PathSeparator)+"findscu")
	command = append(command, "-c", pacs)
	command = append(command, "-b", bind)
	for _, q := range query {
		command = append(command, "-m", q)
	}
	if studyRoot {
		command = append(command, "-r", "00100010") // PatientName
		command = append(command, "-r", "00100020") // PatientID
	}
	command = append(command, "-r", "StudyInstanceUID")
	command = append(command, "-r", "00080050") // AccessionNumber
	command = append(command, "-r", "00080061") // ModalitiesInStudy
	command = append(command, "-r", "00201206") // NumberOfStudyRelatedSeries
	command = append(command, "-r", "00201208") // NumberOfStudyRelatedInstances
	command = append(command, "-r", "00080090") // ReferringPhysicianName
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
	rnPath := xmlpath.MustCompile("DicomAttribute[@keyword='ReferringPhysicianName']/PersonName[@number='1']/Alphabetic/FamilyName")
	pnPath := xmlpath.MustCompile("DicomAttribute[@keyword='PatientName']/PersonName[@number='1']/Alphabetic/FamilyName")
	pIDPath := xmlpath.MustCompile("DicomAttribute[@keyword='PatientID']/Value")
	iter := path.Iter(root)
	for iter.Next() {
		suid, _ := suidPath.String(iter.Node())
		san, _ := sanPath.String(iter.Node())
		smod, _ := smodPath.String(iter.Node())
		sNSer, _ := sNSerPath.String(iter.Node())
		sNIns, _ := sNInsPath.String(iter.Node())
		rn, _ := rnPath.String(iter.Node())
		pn, _ := pnPath.String(iter.Node())
		pID, _ := pIDPath.String(iter.Node())
		csl := tag.StudyLevel{StudyInstanceUID: suid,
			AccessionNumber:          san,
			ModalitiesInStudy:        smod,
			ReferringPhysicianName:   rn,
			NumberOfRelatedSeries:    sNSer,
			NumberOfRelatedInstances: sNIns,
			PatientLevel:             tag.PatientLevel{},
		}
		csl.PatientLevel.PatientName = pn
		csl.PatientLevel.PatientID = pID
		debugf("%v\n", csl)
		sl = append(sl, csl)
	}
	sort.Stable(byStudyInstanceUID(sl))
	return sl, nil
}

func seriesList(bin, pacs, bind, dir, patient, studyUID string) ([]tag.SeriesLevel, error) {
	var sl []tag.SeriesLevel
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
			tag.SeriesLevel{SeriesInstanceUID: suid,
				SeriesNumber:             sn,
				Modality:                 smod,
				NumberOfRelatedInstances: sNIns})
		debugf("%v\n", sl)
	}
	sort.Stable(bySeriesNumber(sl))
	return sl, nil
}

func sopList(bin, pacs, bind, dir, patient, studyUID, seriesUID string) ([]tag.InstanceLevel, error) {
	var sl []tag.InstanceLevel
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
		sl = append(sl, tag.InstanceLevel{SOPInstanceUID: sop, Modality: modality, InstanceNumber: number})
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

func printPatientSOPList(bin, pacs, bind, dir string, level int, get bool, query ...string) error {
	pl, err := patientLevelFind(bin, pacs, bind, dir, query...)
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
			printStudySOPList(bin, pacs, bind, dir, level, get, p, "PatientName="+p.PatientName)
		}
		fmt.Printf("}\n")
	}
	fmt.Printf("patientCount: %d\n", len(pl))
	return nil
}

func printStudySOPList(bin, pacs, bind, dir string, level int, get bool, patient tag.PatientLevel, query ...string) error {
	var studyRoot bool
	if patient.PatientName == "" && patient.PatientID == "" {
		studyRoot = true
	}
	fmt.Printf("  studies: [\n")
	sl, err := studyLevelFind(bin, pacs, bind, dir, studyRoot, query...)
	if err != nil {
		return err
	}
	for _, s := range sl {
		fmt.Printf("    { StudyInstanceUID: %s,\n", s.StudyInstanceUID)
		fmt.Printf("      AccessionNumber: %s,\n", s.AccessionNumber)
		if studyRoot {
			fmt.Printf("      PatientName: %s,\n", s.PatientLevel.PatientName)
			fmt.Printf("      PatientID: %s,\n", s.PatientLevel.PatientID)
		} else {
			s.PatientLevel = patient
		}
		fmt.Printf("      ModalitiesInStudy: %s,\n", s.ModalitiesInStudy)
		fmt.Printf("      NumberOfRelatedSeries: %s,\n", s.NumberOfRelatedSeries)
		fmt.Printf("      NumberOfRelatedInstances: %s,\n", s.NumberOfRelatedInstances)
		fmt.Printf("      ReferringPhysicianName: %s,\n", s.ReferringPhysicianName)
		if level >= 2 { //  series
			fmt.Printf("      series: [\n")
			sel, err := seriesList(bin, pacs, bind, dir, s.PatientLevel.PatientName, s.StudyInstanceUID)
			if err != nil {
				return err
			}
			for _, se := range sel {
				fmt.Printf("        { SeriesInstanceUID: %s,\n", se.SeriesInstanceUID)
				fmt.Printf("          SeriesNumber: %s,\n", se.SeriesNumber)
				fmt.Printf("          Modality: %s,\n", se.Modality)
				fmt.Printf("          NumberOfRelatedInstances: %s,\n", se.NumberOfRelatedInstances)
				if level == 2 && get {
					getSeries(bin, pacs, dir, bind, s.PatientLevel.PatientName, s.StudyInstanceUID, se.SeriesInstanceUID)
				}
				if level >= 3 { //  image
					fmt.Printf("          instances: [\n")
					sopl, err := sopList(bin, pacs, bind, dir, s.PatientLevel.PatientName, s.StudyInstanceUID, se.SeriesInstanceUID)
					if err != nil {
						return err
					}
					if !hideInstances {
						for _, sop := range sopl {
							fmt.Printf("            { Modality: %s, SOPInstanceUID: %s,  InstanceNumber: %s}\n", sop.Modality, sop.SOPInstanceUID, sop.InstanceNumber)
						}
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
	return nil
}

func synopsis() {
	synopsis := `query-retrieve <action> [<patient>]
  --lib <dcm4chee-location> -c|--connect <ae@host:port>
  [--bind <callingAE>]
  [--level <1 study, 2 series, 3 image>] [--hide-instances]
  [--dir <tmp-dir-location>]
  [--debug]

query-retrieve -h # show this help

# actions:
list-all
list-patient <patient>
patient-level-query <patient-level-tag=value>
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
	opt.BoolVar(&hideInstances, "hide-instances", false)
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
		pl, err := patientLevelFind(lib, pacs, bind, dir, "PatientName=*")
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] patientList: %s\n", err)
			os.Exit(1)
		}
		for _, p := range pl {
			err := printPatientSOPList(lib, pacs, bind, dir, level, false, "PatientName="+p.PatientName)
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
		err := printPatientSOPList(lib, pacs, bind, dir, level, false, "PatientName="+patient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] printPatientSOPList: %s\n", err)
			os.Exit(1)
		}
	case "patient-level-query":
		if len(remaining) < 2 {
			fmt.Printf("[ERROR] Missing query!\n")
			synopsis()
			os.Exit(1)
		}
		query := remaining[1:]
		err := printPatientSOPList(lib, pacs, bind, dir, level, false, query...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] printPatientSOPList: %s\n", err)
			os.Exit(1)
		}
	case "study-level-query":
		if len(remaining) < 2 {
			fmt.Printf("[ERROR] Missing query!\n")
			synopsis()
			os.Exit(1)
		}
		query := remaining[1:]
		err := printStudySOPList(lib, pacs, bind, dir, level, false, tag.PatientLevel{}, query...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] printStudySOPList: %s\n", err)
			os.Exit(1)
		}
	case "get-patient":
		if len(remaining) < 2 {
			fmt.Printf("[ERROR] Missing patient!\n")
			synopsis()
			os.Exit(1)
		}
		patient := remaining[1]
		err := printPatientSOPList(lib, pacs, bind, dir, 2, true, "PatientName="+patient)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] printPatientSOPList: %s\n", err)
			os.Exit(1)
		}
	}
}
