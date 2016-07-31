package soap

import (
	"github.com/masterzen/simplexml/dom"
	. "gopkg.in/check.v1"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestNewHeader(c *C) {
	ho := NewHeaderOption("WINRS_CODEPAGE", "65001")
	c.Assert(ho.key, Equals, "WINRS_CODEPAGE")
	c.Assert(ho.value, Equals, "65001")
}

func initDocument() (h *SoapHeader) {
	doc := dom.CreateDocument()
	doc.PrettyPrint = true
	e := dom.CreateElement("Envelope")
	doc.SetRoot(e)
	AddUsualNamespaces(e)
	NS_SOAP_ENV.SetTo(e)
	h = &SoapHeader{message: &SoapMessage{document: doc, envelope: e}}
	return
}

func (s *MySuite) TestHeaderBuild(c *C) {
	h := initDocument()
	msg := h.To("http://winrm:5985/wsman").ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").MaxEnvelopeSize(153600).Id("1-2-3-4").Locale("en_US").Timeout("PT60S").
		Action("http://schemas.xmlsoap.org/ws/2004/09/transfer/Create").Build()

	expected := `<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
  <env:Header>
    <a:To>http://winrm:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <a:MessageID>1-2-3-4</a:MessageID>
    <w:Locale mustUnderstand="false" xml:lang="en_US"/>
    <p:DataLocale mustUnderstand="false" xml:lang="en_US"/>
    <a:Action mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/Create</a:Action>
  </env:Header>
</env:Envelope>
`

	c.Check(msg.String(), Equals, expected)
}

func (s *MySuite) TestOtherHeaderBuild(c *C) {
	h := initDocument()
	msg := h.To("http://winrm:5985/wsman").ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").MaxEnvelopeSize(153600).Id("1-2-3-4").Locale("en_US").Timeout("PT60S").Action("http://schemas.xmlsoap.org/ws/2004/09/transfer/Delete").ShellId("shell-id").ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").Build()

	expected := `<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
  <env:Header>
    <a:To>http://winrm:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <a:MessageID>1-2-3-4</a:MessageID>
    <w:Locale mustUnderstand="false" xml:lang="en_US"/>
    <p:DataLocale mustUnderstand="false" xml:lang="en_US"/>
    <a:Action mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/09/transfer/Delete</a:Action>
    <w:SelectorSet>
      <w:Selector Name="ShellId">shell-id</w:Selector>
    </w:SelectorSet>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
  </env:Header>
</env:Envelope>
`

	c.Check(msg.String(), Equals, expected)
}

func (s *MySuite) TestAddOptionHeaderBuild(c *C) {
	h := initDocument()
	msg := h.To("http://winrm:5985/wsman").ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").MaxEnvelopeSize(153600).Id("1-2-3-4").Locale("en_US").Timeout("PT60S").Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command").ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").ShellId("shell-id").AddOption(NewHeaderOption("WINRS_CONSOLEMODE_STDIN", "TRUE")).AddOption(NewHeaderOption("WINRS_SKIP_CMD_SHELL", "FALSE")).Build()

	expected := `<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
  <env:Header>
    <a:To>http://winrm:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <a:MessageID>1-2-3-4</a:MessageID>
    <w:Locale mustUnderstand="false" xml:lang="en_US"/>
    <p:DataLocale mustUnderstand="false" xml:lang="en_US"/>
    <a:Action mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command</a:Action>
    <w:SelectorSet>
      <w:Selector Name="ShellId">shell-id</w:Selector>
    </w:SelectorSet>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:OptionSet>
      <w:Option Name="WINRS_CONSOLEMODE_STDIN">TRUE</w:Option>
      <w:Option Name="WINRS_SKIP_CMD_SHELL">FALSE</w:Option>
    </w:OptionSet>
  </env:Header>
</env:Envelope>
`

	c.Check(msg.String(), Equals, expected)
}

func (s *MySuite) TestOptionsHeaderBuild(c *C) {
	h := initDocument()
	msg := h.To("http://winrm:5985/wsman").ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").MaxEnvelopeSize(153600).Id("1-2-3-4").Locale("en_US").Timeout("PT60S").
		Action("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command").ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").ShellId("shell-id").
		Options([]HeaderOption{*NewHeaderOption("WINRS_CONSOLEMODE_STDIN", "TRUE"), *NewHeaderOption("WINRS_SKIP_CMD_SHELL", "FALSE")}).Build()

	expected := `<?xml version="1.0" encoding="utf-8" ?>
<env:Envelope xmlns:env="http://www.w3.org/2003/05/soap-envelope" xmlns:a="http://schemas.xmlsoap.org/ws/2004/08/addressing" xmlns:rsp="http://schemas.microsoft.com/wbem/wsman/1/windows/shell" xmlns:w="http://schemas.dmtf.org/wbem/wsman/1/wsman.xsd" xmlns:p="http://schemas.microsoft.com/wbem/wsman/1/wsman.xsd">
  <env:Header>
    <a:To>http://winrm:5985/wsman</a:To>
    <a:ReplyTo>
      <a:Address mustUnderstand="true">http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</a:Address>
    </a:ReplyTo>
    <w:MaxEnvelopeSize mustUnderstand="true">153600</w:MaxEnvelopeSize>
    <w:OperationTimeout>PT60S</w:OperationTimeout>
    <a:MessageID>1-2-3-4</a:MessageID>
    <w:Locale mustUnderstand="false" xml:lang="en_US"/>
    <p:DataLocale mustUnderstand="false" xml:lang="en_US"/>
    <a:Action mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/Command</a:Action>
    <w:SelectorSet>
      <w:Selector Name="ShellId">shell-id</w:Selector>
    </w:SelectorSet>
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:OptionSet>
      <w:Option Name="WINRS_CONSOLEMODE_STDIN">TRUE</w:Option>
      <w:Option Name="WINRS_SKIP_CMD_SHELL">FALSE</w:Option>
    </w:OptionSet>
  </env:Header>
</env:Envelope>
`

	c.Check(msg.String(), Equals, expected)
}
