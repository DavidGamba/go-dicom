// This file is part of go-dicom.
//
// Copyright (C) 2016  David Gamba Rios
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package tag has structures for DICOM tags at the different levels
package tag

// PatientLevel has patient level DICOM tags
type PatientLevel struct {
	PatientName              string // 	(0010,0010)
	PatientID                string //	(0010,0020)
	PatientBirthDate         string //	(0010,0030)
	NumberOfRelatedStudies   string //	(0020,1200)
	NumberOfRelatedSeries    string //	(0020,1202)
	NumberOfRelatedInstances string //	(0020,1204)
}

// StudyLevel has study level DICOM tags
type StudyLevel struct {
	StudyInstanceUID         string //	(0020,000D)
	StudyDate                string //	(0008,0020)
	AccessionNumber          string //	(0008,0050)
	ModalitiesInStudy        string //	(0008,0061)
	ReferringPhysicianName   string //  (0008,0090)
	StudyDescription         string //	(0008,1030)
	NumberOfRelatedSeries    string //	(0020,1206)
	NumberOfRelatedInstances string //	(0020,1208)
}

// SeriesLevel has series level DICOM tags
type SeriesLevel struct {
	SeriesInstanceUID        string //	(0020,000E)
	SeriesNumber             string //	(0020,0011)
	Modality                 string //	(0008,0060)
	NumberOfRelatedInstances string //	(0020,1209)
}

// InstanceLevel has image level DICOM tags
type InstanceLevel struct {
	SOPInstanceUID string //	(0000,0000)
	Modality       string //	(0000,0000)
	InstanceNumber string //	(0000,0000)
}
