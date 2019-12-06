package approvaltests

import (
	"fmt"
	"strings"

	"reflect"
)

type emptyType struct{}

var (
	// SkipThisCombination should be returned if you do not want to process a particular combination
	SkipThisCombination = "♬ SKIP THIS COMBINATION ♬"

	empty           = emptyType{}
	emptyCollection = []emptyType{empty}
)

// VerifyAllCombinationsFor1 Example:
//   VerifyAllCombinationsFor1(t, "uppercase", func(x interface{}) string { return strings.ToUpper(x.(string)) }, []string("dog", "cat"})
func VerifyAllCombinationsFor1(t Failable, header string, transform func(interface{}) string, collection1 interface{}) error {
	transform2 := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1)
	}

	return VerifyAllCombinationsFor9(t, header, transform2, collection1,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor2 Example:
//   VerifyAllCombinationsFor2(t, "uppercase", func(x interface{}) string { return strings.ToUpper(x.(string)) }, []string("dog", "cat"}, []int{1,2)
func VerifyAllCombinationsFor2(t Failable, header string, transform func(interface{}, interface{}) string, collection1 interface{}, collection2 interface{}) error {
	transform2 := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2)
	}

	return VerifyAllCombinationsFor9(t, header, transform2, collection1,
		collection2,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor3 is for combinations of 3.
func VerifyAllCombinationsFor3(
	t Failable,
	header string,
	transform func(p1, p2, p3 interface{}) string,
	collection1, collection2, collection3 interface{}) error {

	kerning := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2, p3)
	}

	return VerifyAllCombinationsFor9(t, header, kerning,
		collection1,
		collection2,
		collection3,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor4 is for combinations of 4.
func VerifyAllCombinationsFor4(
	t Failable,
	header string,
	transform func(p1, p2, p3, p4 interface{}) string,
	collection1, collection2, collection3, collection4 interface{}) error {

	kerning := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2, p3, p4)
	}

	return VerifyAllCombinationsFor9(t, header, kerning,
		collection1,
		collection2,
		collection3,
		collection4,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor5 is for combinations of 5.
func VerifyAllCombinationsFor5(
	t Failable,
	header string,
	transform func(p1, p2, p3, p4, p5 interface{}) string,
	collection1, collection2, collection3, collection4, collection5 interface{}) error {

	kerning := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2, p3, p4, p5)
	}

	return VerifyAllCombinationsFor9(t, header, kerning,
		collection1,
		collection2,
		collection3,
		collection4,
		collection5,
		emptyCollection,
		emptyCollection,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor6 is for combinations of 6.
func VerifyAllCombinationsFor6(
	t Failable,
	header string,
	transform func(p1, p2, p3, p4, p5, p6 interface{}) string,
	collection1, collection2, collection3, collection4, collection5, collection6 interface{}) error {

	kerning := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2, p3, p4, p5, p6)
	}

	return VerifyAllCombinationsFor9(t, header, kerning,
		collection1,
		collection2,
		collection3,
		collection4,
		collection5,
		collection6,
		emptyCollection,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor7 is for combinations of 7.
func VerifyAllCombinationsFor7(
	t Failable,
	header string,
	transform func(p1, p2, p3, p4, p5, p6, p7 interface{}) string,
	collection1, collection2, collection3, collection4, collection5, collection6, collection7 interface{}) error {

	kerning := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2, p3, p4, p5, p6, p7)
	}

	return VerifyAllCombinationsFor9(t, header, kerning,
		collection1,
		collection2,
		collection3,
		collection4,
		collection5,
		collection6,
		collection7,
		emptyCollection,
		emptyCollection)
}

// VerifyAllCombinationsFor8 is for combinations of 8.
func VerifyAllCombinationsFor8(
	t Failable,
	header string,
	transform func(p1, p2, p3, p4, p5, p6, p7, p8 interface{}) string,
	collection1, collection2, collection3, collection4, collection5, collection6, collection7, collection8 interface{}) error {

	kerning := func(p1, p2, p3, p4, p5, p6, p7, p8, p9 interface{}) string {
		return transform(p1, p2, p3, p4, p5, p6, p7, p8)
	}

	return VerifyAllCombinationsFor9(t, header, kerning,
		collection1,
		collection2,
		collection3,
		collection4,
		collection5,
		collection6,
		collection7,
		collection8,
		emptyCollection)
}

// VerifyAllCombinationsFor9 is for combinations of 9.
func VerifyAllCombinationsFor9(
	t Failable,
	header string,
	transform func(a, b, c, d, e, f, g, h, i interface{}) string,
	collection1,
	collection2,
	collection3,
	collection4,
	collection5,
	collection6,
	collection7,
	collection8,
	collection9 interface{}) error {

	if len(header) != 0 {
		header = fmt.Sprintf("%s\n\n\n", header)
	}

	var mapped []string

	slice1 := reflect.ValueOf(collection1)
	slice2 := reflect.ValueOf(collection2)
	slice3 := reflect.ValueOf(collection3)
	slice4 := reflect.ValueOf(collection4)
	slice5 := reflect.ValueOf(collection5)
	slice6 := reflect.ValueOf(collection6)
	slice7 := reflect.ValueOf(collection7)
	slice8 := reflect.ValueOf(collection8)
	slice9 := reflect.ValueOf(collection9)

	for i1 := 0; i1 < slice1.Len(); i1++ {
		for i2 := 0; i2 < slice2.Len(); i2++ {
			for i3 := 0; i3 < slice3.Len(); i3++ {
				for i4 := 0; i4 < slice4.Len(); i4++ {
					for i5 := 0; i5 < slice5.Len(); i5++ {
						for i6 := 0; i6 < slice6.Len(); i6++ {
							for i7 := 0; i7 < slice7.Len(); i7++ {
								for i8 := 0; i8 < slice8.Len(); i8++ {
									for i9 := 0; i9 < slice9.Len(); i9++ {
										p1 := slice1.Index(i1).Interface()
										p2 := slice2.Index(i2).Interface()
										p3 := slice3.Index(i3).Interface()
										p4 := slice4.Index(i4).Interface()
										p5 := slice5.Index(i5).Interface()
										p6 := slice6.Index(i6).Interface()
										p7 := slice7.Index(i7).Interface()
										p8 := slice8.Index(i8).Interface()
										p9 := slice9.Index(i9).Interface()

										parameterText := getParameterText(p1, p2, p3, p4, p5, p6, p7, p8, p9)
										transformText := getTransformText(transform, p1, p2, p3, p4, p5, p6, p7, p8, p9)
										if transformText != SkipThisCombination {
											mapped = append(mapped, fmt.Sprintf("%s => %s", parameterText, transformText))
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	outputText := header + strings.Join(mapped, "\n")
	return VerifyString(t, outputText)
}

func getParameterText(args ...interface{}) string {
	parameterText := "["
	for _, x := range args {
		if x != empty {
			parameterText += fmt.Sprintf("%v,", x)
		}
	}

	parameterText = parameterText[0 : len(parameterText)-1]
	parameterText += "]"

	return parameterText
}

func getTransformText(
	transform func(a, b, c, d, e, f, g, h, i interface{}) string,
	p1,
	p2,
	p3,
	p4,
	p5,
	p6,
	p7,
	p8,
	p9 interface{}) (s string) {
	defer func() {
		r := recover()
		if r != nil {
			s = "panic occurred"
		}
	}()

	return transform(p1, p2, p3, p4, p5, p6, p7, p8, p9)
}
