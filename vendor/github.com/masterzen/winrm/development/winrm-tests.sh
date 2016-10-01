#!/bin/bash
echo "==> Running commands using WinRM against a Vagrant machine:"
echo


echo "==> What's my execution policy?"
winrm "powershell -command \"get-executionpolicy\""
echo "==> Whoami?"
winrm "whoami"
