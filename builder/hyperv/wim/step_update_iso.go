package wim

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/hyperv/common/powershell"
)

type StepUpdateISO struct {
	DevicePathKey      string
	ISOPathKey         string
	OriginalISOPathKey string
	SkipOperation      bool
	UseEfiBoot         bool
	WIMPathKey         string
}

func (s *StepUpdateISO) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if s.SkipOperation {
		return multistep.ActionContinue
	}

	ui := state.Get("ui").(packersdk.Ui)
	buildDir := state.Get("build_dir").(string)
	devicePath := state.Get(s.DevicePathKey).(string)
	srcISOPath := state.Get(s.ISOPathKey).(string)
	srcWIMPath := state.Get(s.WIMPathKey).(string)

	ui.Say("Updating ISO...")

	dstDir := filepath.Join(buildDir, "iso")
	dstISOPath := filepath.Join(buildDir, "packer.iso")
	dstWIMPath := filepath.Join(dstDir, installWIMPath)

	// Create a temp folder for ISO content
	if err := os.Mkdir(dstDir, 0777); err != nil {
		err = fmt.Errorf("Error creating an ISO folder: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	defer os.Remove(dstDir)

	devicePath = devicePath + "\\"

	// Copy ISO content to another directory except sources/install.wim
	if err := s.copyDirectories(devicePath, dstDir); err != nil {
		err = fmt.Errorf("Error copy ISO content: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Copied directory from %s to %s", devicePath, dstDir))

	// Replace install.wim
	if err := s.replaceInstallWIM(srcWIMPath, dstWIMPath); err != nil {
		err = fmt.Errorf("Error replacing WIM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Replaced WIM at %s", dstWIMPath))

	// Create ISO

	var bootFile string
	if s.UseEfiBoot {
		bootFile = filepath.Join(dstDir, "efi", "microsoft", "boot", "efisys.bin")
	} else {
		bootFile = filepath.Join(dstDir, "boot", "etfsboot.com")
	}

	ui.Say(fmt.Sprintf("Boot file: %s", bootFile))

	if err := s.createISO(dstDir, dstISOPath, bootFile); err != nil {
		err = fmt.Errorf("Error creating ISO: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Say(fmt.Sprintf("Created ISO at %s", dstISOPath))

	// Update state bag with the new backup ISO
	state.Put(s.OriginalISOPathKey, srcISOPath)
	state.Put(s.ISOPathKey, dstISOPath)

	return multistep.ActionContinue
}

func (s *StepUpdateISO) Cleanup(state multistep.StateBag) {
	if s.SkipOperation {
		return
	}

	ui := state.Get("ui").(packersdk.Ui)

	if originalISOPath, ok := state.GetOk(s.OriginalISOPathKey); ok {
		// Remove ISO
		isoPath := state.Get(s.ISOPathKey).(string)
		_ = os.Remove(isoPath)

		ui.Say(fmt.Sprintf("Removed ISO at %s", isoPath))

		// Revert back to the original state
		state.Put(s.ISOPathKey, originalISOPath)
		state.Remove(s.OriginalISOPathKey)
	}
}

func (s *StepUpdateISO) copyDirectories(srcDir string, dstDir string) error {
	// TODO: except install.wim

	var script = `
param([string]$srcDir,[string]$dstDir)
Get-ChildItem -Path $srcDir | Copy-Item -Destination $dstDir -Recurse -Container
`

	var ps powershell.PowerShellCmd
	return ps.Run(script, srcDir, dstDir)
}

func (s *StepUpdateISO) createISO(path string, isoPath string, bootFile string) error {

	// Modified from https://gallery.technet.microsoft.com/scriptcenter/New-ISOFile-function-a8deeffd
	// This probably won't work on Windows Server Core, so may need a different solution later
	var script = `
function New-IsoFile
{
  <#
   .Synopsis
    Creates a new .iso file
   .Description
    The New-IsoFile cmdlet creates a new .iso file containing content from chosen folders
   .Example
    New-IsoFile "c:\tools","c:Downloads\utils"
    This command creates a .iso file in $env:temp folder (default location) that contains c:\tools and c:\downloads\utils folders. The folders themselves are included at the root of the .iso image.
   .Example
    New-IsoFile -FromClipboard -Verbose
    Before running this command, select and copy (Ctrl-C) files/folders in Explorer first.
   .Example
    dir c:\WinPE | New-IsoFile -Path c:\temp\WinPE.iso -BootFile "${env:ProgramFiles(x86)}\Windows Kits\10\Assessment and Deployment Kit\Deployment Tools\amd64\Oscdimg\efisys.bin" -Media DVDPLUSR -Title "WinPE"
    This command creates a bootable .iso file containing the content from c:\WinPE folder, but the folder itself isn't included. Boot file etfsboot.com can be found in Windows ADK. Refer to IMAPI_MEDIA_PHYSICAL_TYPE enumeration for possible media types: http://msdn.microsoft.com/en-us/library/windows/desktop/aa366217(v=vs.85).aspx
   .Notes
    NAME:  New-IsoFile
    AUTHOR: Chris Wu
    LASTEDIT: 03/23/2016 14:46:50
 #>

    [CmdletBinding(DefaultParameterSetName='Source')]Param(
    [parameter(Position=1,Mandatory=$true,ValueFromPipeline=$true, ParameterSetName='Source')]$Source,
    [parameter(Position=2)][string]$Path = "$env:temp\$((Get-Date).ToString('yyyyMMdd-HHmmss.ffff')).iso",
    [ValidateScript({Test-Path -LiteralPath $_ -PathType Leaf})][string]$BootFile = $null,
    [ValidateSet('CDR','CDRW','DVDRAM','DVDPLUSR','DVDPLUSRW','DVDPLUSR_DUALLAYER','DVDDASHR','DVDDASHRW','DVDDASHR_DUALLAYER','DISK','DVDPLUSRW_DUALLAYER','BDR','BDRE')][string] $Media = 'DVDPLUSRW_DUALLAYER',
    [string]$Title = (Get-Date).ToString("yyyyMMdd-HHmmss.ffff"),
    [switch]$Force,
    [parameter(ParameterSetName='Clipboard')][switch]$FromClipboard
  )

  Begin {
    ($cp = new-object System.CodeDom.Compiler.CompilerParameters).CompilerOptions = '/unsafe'
    if (!('ISOFile' -as [type])) {
      Add-Type -CompilerParameters $cp -TypeDefinition @'
public class ISOFile
{
  public unsafe static void Create(string Path, object Stream, int BlockSize, int TotalBlocks)
  {
    int bytes = 0;
    byte[] buf = new byte[BlockSize];
    var ptr = (System.IntPtr)(&bytes);
    var o = System.IO.File.OpenWrite(Path);
    var i = Stream as System.Runtime.InteropServices.ComTypes.IStream;

    if (o != null) {
      while (TotalBlocks-- > 0) {
        i.Read(buf, BlockSize, ptr); o.Write(buf, 0, bytes);
      }
      o.Flush(); o.Close();
    }
  }
}
'@
    }

    if ($BootFile) {
      if('BDR','BDRE' -contains $Media) { Write-Warning "Bootable image doesn't seem to work with media type $Media" }
      ($Stream = New-Object -ComObject ADODB.Stream -Property @{Type=1}).Open()  # adFileTypeBinary
      $Stream.LoadFromFile((Get-Item -LiteralPath $BootFile).Fullname)
      ($Boot = New-Object -ComObject IMAPI2FS.BootOptions).AssignBootImage($Stream)
    }

    $MediaType = @('UNKNOWN','CDROM','CDR','CDRW','DVDROM','DVDRAM','DVDPLUSR','DVDPLUSRW','DVDPLUSR_DUALLAYER','DVDDASHR','DVDDASHRW','DVDDASHR_DUALLAYER','DISK','DVDPLUSRW_DUALLAYER','HDDVDROM','HDDVDR','HDDVDRAM','BDROM','BDR','BDRE')

    Write-Verbose -Message "Selected media type is $Media with value $($MediaType.IndexOf($Media))"
    ($Image = New-Object -com IMAPI2FS.MsftFileSystemImage -Property @{VolumeName=$Title}).ChooseImageDefaultsForMediaType($MediaType.IndexOf($Media))

    if (!($Target = New-Item -Path $Path -ItemType File -Force:$Force -ErrorAction SilentlyContinue)) { Write-Error -Message "Cannot create file $Path. Use -Force parameter to overwrite if the target file already exists."; break }
  }

  Process {
    if($FromClipboard) {
      if($PSVersionTable.PSVersion.Major -lt 5) { Write-Error -Message 'The -FromClipboard parameter is only supported on PowerShell v5 or higher'; break }
      $Source = Get-Clipboard -Format FileDropList
    }

    $items = Get-ChildItem -Path $Source
    foreach ($item in $items) {
      if($item -isnot [System.IO.FileInfo] -and $item -isnot [System.IO.DirectoryInfo]) {
        $item = Get-Item -LiteralPath $item
      }

      if($item) {
        Write-Information -Message "Adding item to the target image: $($item.FullName)"
        try { $Image.Root.AddTree($item.FullName, $true) } catch { Write-Error -Message ($_.Exception.Message.Trim() + ' Try a different media type.') }
      }
    }
  }

  End {
    if ($Boot) { $Image.BootImageOptions=$Boot }
    $Result = $Image.CreateResultImage()
    [ISOFile]::Create($Target.FullName,$Result.ImageStream,$Result.BlockSize,$Result.TotalBlocks)
    Write-Verbose -Message "Target image ($($Target.FullName)) has been created"
    $Target
  }
}

`

	script = fmt.Sprintf("%sNew-IsoFile -Path %s -Source %s\\ -BootFile %s -Force", script, isoPath, path, bootFile)

	var ps powershell.PowerShellCmd
	return ps.Run(script)
}

func (s *StepUpdateISO) replaceInstallWIM(srcWIMPath string, dstWIMPath string) error {

	srcWIM, err := os.Open(srcWIMPath)
	if err != nil {
		return err
	}

	defer srcWIM.Close()

	if err = os.Remove(dstWIMPath); err != nil {
		return err
	}

	var dstWIM *os.File
	if dstWIM, err = os.Create(dstWIMPath); err != nil {
		return err
	}

	// dstWIM, err := os.OpenFile(dstWIMPath, os.O_WRONLY, 0666)
	// if err != nil {
	// 	return err
	// }

	defer dstWIM.Close()

	_, err = io.Copy(dstWIM, srcWIM)
	if err != nil {
		return err
	}

	return nil
}
