package soap

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestNewMessage(c *C) {
	message := NewMessage()
	defer message.Free()
	message.Doc().PrettyPrint = true
	message.Header().To("http://winrm:5985/wsman").
		ReplyTo("http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous").
		MaxEnvelopeSize(153600).Id("1-2-3-4").Locale("en_US").Timeout("PT60S").
		Action("http://schemas.xmlsoap.org/ws/2004/09/transfer/Create").
		ResourceURI("http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd").
		AddOption(NewHeaderOption("WINRS_NOPROFILE", "FALSE")).
		AddOption(NewHeaderOption("WINRS_CODEPAGE", "65001")).
		Build()

	body := message.CreateBodyElement("Shell", NS_WIN_SHELL)
	input := message.CreateElement(body, "InputStreams", NS_WIN_SHELL)
	input.SetContent("stdin")
	output := message.CreateElement(body, "OutputStreams", NS_WIN_SHELL)
	output.SetContent("stdout stderr")

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
    <w:ResourceURI mustUnderstand="true">http://schemas.microsoft.com/wbem/wsman/1/windows/shell/cmd</w:ResourceURI>
    <w:OptionSet>
      <w:Option Name="WINRS_NOPROFILE">FALSE</w:Option>
      <w:Option Name="WINRS_CODEPAGE">65001</w:Option>
    </w:OptionSet>
  </env:Header>
  <env:Body>
    <rsp:Shell>
      <rsp:InputStreams>stdin</rsp:InputStreams>
      <rsp:OutputStreams>stdout stderr</rsp:OutputStreams>
    </rsp:Shell>
  </env:Body>
</env:Envelope>
`

	c.Check(message.String(), Equals, expected)
}
