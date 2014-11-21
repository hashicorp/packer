// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import (
	"encoding/xml"
	"regexp"
	"fmt"
	"bytes"
	"strings"
	"sort"
)

type VmImageList struct {
	XMLName   xml.Name `xml:"VMImages"`
	Xmlns	  	string `xml:"xmlns,attr"`
	VMImages []VMImage `xml:"VMImage"`
}

type VMImage struct {
	Name  						string
	Label  						string
	Category  					string
	Description  				string
	OSDiskConfiguration  		OSDiskConfiguration
	DataDiskConfigurations		[]DataDiskConfiguration	`xml:"DataDiskConfigurations > DataDiskConfiguration"`
	ServiceName					string
	DeploymentName				string
	RoleName					string
	Location					string
	AffinityGroup				string
	CreatedTime					string
	ModifiedTime				string
	Language					string
	ImageFamily					string
	RecommendedVMSize			string
	IsPremium					string
	Eula						string
	IconUri						string
	SmallIconUri				string
	PrivacyUri					string
	PublisherName				string
	PublishedDate				string
	ShowInGui					string
	PricingDetailLink			string
}

type OSDiskConfiguration struct {
	Name  					string
	HostCaching  			string
	OSState  				string
	OS  					string
	MediaLink  				string
	LogicalDiskSizeInGB  	string
}

type DataDiskConfiguration struct {
	Name  					string
	HostCaching  			string
	Lun  					string
	MediaLink  				string
	LogicalDiskSizeInGB  	string
}


func (l *VmImageList) First(name string) *VMImage {
	pattern := name
	for _, im := range(l.VMImages){
		matchName, _ := regexp.MatchString(pattern, im.Name)
		if( matchName ) {
			return &im
		}
	}

	return nil
}

func (l *VmImageList) Filter(label, location string) []VMImage {
	dgb_output := false

	origLen := len(l.VMImages)
	filtered  := make([]VMImage, 0, origLen)

	var matchImageLabel bool
	var matchImageFamily bool

	for _, im := range(l.VMImages) {

		if im.OSDiskConfiguration.OSState != "Generalized" {
			continue
		}

		matchImageLocation := false
		for _, loc := range strings.Split(im.Location, ";")	{
			if loc == location {
				matchImageLocation = true;
				break
			}
		}
		if !matchImageLocation { continue }

		matchImageLabel = strings.Contains(im.Label, label)
		matchImageFamily =  strings.Contains(im.ImageFamily, label)

		if dgb_output {
			fmt.Printf("label: '%s'\nfamily: '%s'\nlocations: '%s'\nPublishedDate: '%s'\nl-f: '%v-%v'\n\n",
				im.Label, im.ImageFamily, im.Location, im.PublishedDate, matchImageLabel, matchImageFamily)
		}

		if matchImageLabel || matchImageFamily  {
			filtered = append( filtered, im)
		}
	}

	return filtered[:len(filtered)]
}

type VMImageByDateDesc []VMImage

func (a VMImageByDateDesc) Len() int           { return len(a) }
func (a VMImageByDateDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a VMImageByDateDesc) Less(i, j int) bool {
	if len(a[i].PublishedDate) >0 &&  len(a[j].PublishedDate) >0 {
		return a[i].PublishedDate > a[j].PublishedDate
	}
	// assumed name contains creation date
	return a[i].CreatedTime > a[j].CreatedTime
}

func (l *VmImageList) SortByDateDesc(images []VMImage) {
	if len(images) == 0 {
		sort.Sort(VMImageByDateDesc(l.VMImages))
	} else {
		sort.Sort(VMImageByDateDesc(images))
	}
}

func PrintVmImages(images []VMImage){
	var output bytes.Buffer
	for _, im := range(images) {
		output.Reset()
		output.WriteString("Label: " + 			im.Label 			+ "\n")
		output.WriteString("Location: " + 		im.Location 		+ "\n")
		output.WriteString("Name: " + 			im.Name 			+ "\n")
		output.WriteString("ImageFamily: " + 	im.ImageFamily 		+ "\n")
		output.WriteString("PublishedDate: " +	im.PublishedDate 	+ "\n")
		output.WriteString("CreatedTime: " +	im.CreatedTime 	+ "\n")
		output.WriteString("ModifiedTime: " +	im.ModifiedTime 	+ "\n")
		output.WriteString("Category: " + 		im.Category 		+ "\n")
		output.WriteString("OsState: " + 		im.OSDiskConfiguration.OSState 		+ "\n")
		output.WriteString("MediaLink: " + 		im.OSDiskConfiguration.MediaLink 	+ "\n")

		fmt.Println(output.String())
	}
}

