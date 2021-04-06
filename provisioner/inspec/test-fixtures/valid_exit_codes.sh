#!/bin/sh

cat <<EOB
Profile: tests from test-fixtures/skip_control.rb (tests from test-fixtures/valid_exit_codes.sh)
Version: (not specified)
Target:  local://

[38;5;247m  ↺  skip-1.0: skip control[0m
[38;5;247m     ↺  Skipped control due to only_if condition.[0m


Profile Summary: 0 successful controls, 0 control failures, [38;5;247m1 control skipped[0m
Test Summary: 0 successful, 0 failures, [38;5;247m1 skipped[0m
EOB

exit 101
