package googlecompute

import (
	"encoding/base64"
	"fmt"
)

const StartupScriptStartLog string = "Packer startup script starting."
const StartupScriptDoneLog string = "Packer startup script done."
const StartupScriptKey string = "startup-script"
const StartupWrappedScriptKey string = "packer-wrapped-startup-script"

// We have to encode StartupScriptDoneLog because we use it as a sentinel value to indicate
// that the user-provided startup script is done. If we pass StartupScriptDoneLog as-is, it
// will be printed early in the instance console log (before the startup script even runs;
// we print out instance creation metadata which contains this wrapper script).
var StartupScriptDoneLogBase64 string = base64.StdEncoding.EncodeToString([]byte(StartupScriptDoneLog))

var StartupScript string = fmt.Sprintf(`#!/bin/bash
echo %s
RETVAL=0

GetMetadata () {
	echo "$(curl -f -H "Metadata-Flavor: Google" http://metadata/computeMetadata/v1/instance/attributes/$1 2> /dev/null)"
}

STARTUPSCRIPT=$(GetMetadata %s)
STARTUPSCRIPTPATH=/packer-wrapped-startup-script
if [ -f "/var/log/startupscript.log" ]; then
  STARTUPSCRIPTLOGPATH=/var/log/startupscript.log
else
  STARTUPSCRIPTLOGPATH=/var/log/daemon.log
fi
STARTUPSCRIPTLOGDEST=$(GetMetadata startup-script-log-dest)

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

echo $(echo %s | base64 --decode)
exit $RETVAL
`, StartupScriptStartLog, StartupWrappedScriptKey, StartupScriptDoneLogBase64)
