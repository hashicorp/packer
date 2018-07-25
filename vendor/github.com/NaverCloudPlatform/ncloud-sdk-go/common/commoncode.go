package common

type CommonCode struct {
	CodeKind             string `xml:"codeKind"`
	DetailCategorizeCode string `xml:"detailCategorizeCode"`
	Code                 string `xml:"code"`
	CodeName             string `xml:"codeName"`
	CodeOrder            int    `xml:"codeOrder"`
	JavaConstantCode     string `xml:"javaConstantCode"`
}
