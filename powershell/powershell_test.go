package powershell

import (
	"bytes"
	"testing"
)

func TestOutput(t *testing.T) {
	var ps PowerShellCmd

	powerShellPath, err := ps.getPowerShellPath()
	if err != nil {
		t.Skipf("powershell not installed: %s", err)
		return
	}

	cmdOut, err := ps.Output("")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if cmdOut != "" {
		t.Fatalf("output '%v' is not ''", cmdOut)
	}

	trueOutput, err := ps.Output("$True")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if trueOutput != "True" {
		t.Fatalf("output '%v' is not 'True'", trueOutput)
	}

	falseOutput, err := ps.Output("$False")
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if falseOutput != "False" {
		t.Fatalf("output '%v' is not 'False'", falseOutput)
	}
}

func TestRunFile(t *testing.T) {
	var blockBuffer bytes.Buffer
	blockBuffer.WriteString(`param([string]$a, [string]$b, [int]$x, [int]$y) if (Test-Path variable:global:ProgressPreference){$ProgressPreference="SilentlyContinue"}; $n = $x + $y; Write-Output "$a $b $n";`)

	var ps PowerShellCmd
	
	powerShellPath, err := ps.getPowerShellPath()
	if err != nil {
		t.Skipf("powershell not installed: %s", err)
		return
	}
	
	cmdOut, err := ps.Output(blockBuffer.String(), "a", "b", "5", "10")

	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if cmdOut != "a b 15" {
		t.Fatalf("output '%v' is not 'a b 15'", cmdOut)
	}
}
