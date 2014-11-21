// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package model

import (
	"encoding/xml"
//	"regexp"
	"sort"
	"strings"
	"fmt"
	"bytes"
)

type OsImageList struct {
	XMLName   xml.Name `xml:"Images"`
	Xmlns	  	string `xml:"xmlns,attr"`
	OSImages []OSImage `xml:"OSImage"`
}

type OSImage struct {
	AffinityGroup  		string
	Category			string
	Label				string
	Location			string
	LogicalSizeInGB		string
	MediaLink			string
	Name				string
	OS					string
	Eula				string
	Description			string
	ImageFamily			string
	ShowInGui			string
	PublishedDate		string
	IsPremium			string
	PrivacyUri			string
	RecommendedVMSize	string
	PublisherName		string
	PricingDetailLink	string
	SmallIconUri		string
	Language			string
}

// ignore daily built images (example: Ubuntu Server 14.04.1 LTS DAILY)
const ignoreDAILY bool = true

func (l *OsImageList) Filter(label, location string) []OSImage {
	dgb_output := false

	origLen := len(l.OSImages)
	filtered  := make([]OSImage, 0, origLen)

	var matchImageLabel bool
	var matchImageFamily bool

	for _, im := range(l.OSImages) {

		if ignoreDAILY && strings.Contains(im.Label, "DAILY") {
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

type OSImageByDateDesc []OSImage

func (a OSImageByDateDesc) Len() int           { return len(a) }
func (a OSImageByDateDesc) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a OSImageByDateDesc) Less(i, j int) bool { return a[i].PublishedDate > a[j].PublishedDate }

func (l *OsImageList) SortByDateDesc(images []OSImage) {
	if len(images) == 0 {
		sort.Sort(OSImageByDateDesc(l.OSImages))
	} else {
		sort.Sort(OSImageByDateDesc(images))
	}
}

func PrintOsImages(images []OSImage){
	var output bytes.Buffer
	for _, im := range(images) {
		output.Reset()
		output.WriteString("Label: " + 			im.Label 			+ "\n")
		output.WriteString("Location: " + 		im.Location 		+ "\n")
		output.WriteString("Name: " + 			im.Name 			+ "\n")
		output.WriteString("ImageFamily: " + 	im.ImageFamily 		+ "\n")
		output.WriteString("PublishedDate: " +	im.PublishedDate 	+ "\n")
		output.WriteString("Category: " + 		im.Category 		+ "\n")
		output.WriteString("MediaLink: " + 		im.MediaLink 		+ "\n")

		fmt.Println(output.String())
	}
}
