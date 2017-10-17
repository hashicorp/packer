package googlecompute

import (
	"fmt"
)

const StartupScriptKey string = "startup-script"
const StartupScriptStatusKey string = "startup-script-status"
const StartupWrappedScriptKey string = "packer-wrapped-startup-script"

const StartupScriptStatusDone string = "done"
const StartupScriptStatusError string = "error"
const StartupScriptStatusNotDone string = "notdone"

var StartupScriptLinux string = fmt.Sprintf(`#!/usr/bin/env bash
echo "Packer startup script starting."
RETVAL=0
BASEMETADATAURL=http://metadata/computeMetadata/v1/instance/

GetMetadata () {
  echo "$(curl -f -H "Metadata-Flavor: Google" ${BASEMETADATAURL}/${1} 2> /dev/null)"
}

ZONE=$(basename $(GetMetadata zone))

SetMetadata () {
  gcloud compute instances add-metadata ${HOSTNAME} --metadata ${1}=${2} --zone ${ZONE}
}

STARTUPSCRIPT=$(GetMetadata attributes/%s)
STARTUPSCRIPTPATH=/packer-wrapped-startup-script
if [ -f "/var/log/startupscript.log" ]; then
  STARTUPSCRIPTLOGPATH=/var/log/startupscript.log
else
  STARTUPSCRIPTLOGPATH=/var/log/daemon.log
fi
STARTUPSCRIPTLOGDEST=$(GetMetadata attributes/startup-script-log-dest)

if [[ ! -z $STARTUPSCRIPT ]]; then
  echo "Executing user-provided startup script..."
  echo "${STARTUPSCRIPT}" > ${STARTUPSCRIPTPATH}
  chmod +x ${STARTUPSCRIPTPATH}
  ${STARTUPSCRIPTPATH}
  RETVAL=$?

  if [[ ! -z $STARTUPSCRIPTLOGDEST ]]; then
    echo "Uploading user-provided startup script log to ${STARTUPSCRIPTLOGDEST}..."
    gsutil -h "Content-Type:text/plain" cp ${STARTUPSCRIPTLOGPATH} ${STARTUPSCRIPTLOGDEST}
  fi

  rm ${STARTUPSCRIPTPATH}
fi

echo "Packer startup script done."
SetMetadata %s %s
exit $RETVAL
`, StartupWrappedScriptKey, StartupScriptStatusKey, StartupScriptStatusDone)

var StartupScriptWindows string = ""
