

package powershell

import (
	"bytes"
	"testing"
)

func TestOutputScriptBlock(t *testing.T) {

	ps, err := powershell.Command()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	trueOutput, err := powershell.OutputScriptBlock("$True")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if trueOutput != "True" {
		t.Fatalf("output '%v' is not 'True'", trueOutput)
	}

	falseOutput, err := powershell.OutputScriptBlock("$False")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if falseOutput != "False" {
		t.Fatalf("output '%v' is not 'False'", falseOutput)
	}
}

func TestRunScriptBlock(t *testing.T) {
	powershell, err := powershell.Command()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	err = powershell.RunScriptBlock("$True")
}

func TestVersion(t *testing.T) {
	powershell, err := powershell.Command()
	version, err := powershell.Version();
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if (version != 4) {
		t.Fatalf("expected version 4")
	}
}

func TestRunFile(t *testing.T) {
	powershell, err := powershell.Command()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	var blockBuffer bytes.Buffer
	blockBuffer.WriteString("param([string]$a, [string]$b, [int]$x, [int]$y) $n = $x + $y; Write-Host $a, $b, $n")

	err = powershell.Run(blockBuffer.String(), "a", "b", "5", "10")

	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

}
