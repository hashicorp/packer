# go-select

select(2) implementation in Go

## Supported platforms

|               | 386 | amd64 | arm | arm64 | mips | mipsle | mips64 | mips64le | ppc64le | s390x |
|---------------|-----|-------|-----|-------|------|--------|--------|----------|---------|-------|
| **linux**     | yes | yes   | yes | yes   | yes  | yes    | yes    | yes      | yes     | yes   |
| **darwin**    | yes | yes   | ??  | ??    | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **freebsd**   | yes | yes   | yes | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **openbsd**   | yes | yes   | yes | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **netbsd**    | yes | yes   | yes | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **dragonfly** | n/a | yes   | n/a | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **solaris**   | n/a | no    | n/a | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **plan9**     | no  | no    | no  | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **windows**   | yes | yes   | n/a | n/a   | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |
| **android**   | ??  | ??    | ??  | ??    | n/a  | n/a    | n/a    | n/a      | n/a     | n/a   |

*n/a: platform not supported by Go
*??: not tested

Go on `plan9` and `solaris` do not implement `syscall.Select` nor `syscall.SYS_SELECT`.

## Cross compile test

Note that this only tests the compilation, not the functionality.

```sh
$> ./test_crosscompile.sh > /dev/null | sort
[OK] android/386
[OK] android/amd64
[OK] android/arm
[OK] android/arm64
[OK] darwin/386
[OK] darwin/amd64
[OK] darwin/arm
[OK] darwin/arm64
[OK] dragonfly/amd64
[OK] freebsd/386
[OK] freebsd/amd64
[OK] freebsd/arm
[OK] linux/386
[OK] linux/amd64
[OK] linux/arm
[OK] linux/arm64
[OK] linux/mips
[OK] linux/mips64
[OK] linux/mips64le
[OK] linux/mipsle
[OK] linux/ppc64le
[OK] linux/s390x
[OK] netbsd/386
[OK] netbsd/amd64
[OK] netbsd/arm
[OK] openbsd/386
[OK] openbsd/amd64
[OK] openbsd/arm
[OK] plan9/386
[OK] plan9/amd64
[OK] plan9/arm
[OK] solaris/amd64
[OK] windows/386
[OK] windows/amd64
[OK] windows/arm

# Expected failures.
[KO] android/mips
[KO] android/mips64
[KO] android/mips64le
[KO] android/mipsle
[KO] android/ppc64le
[KO] android/s390x
[KO] darwin/mips
[KO] darwin/mips64
[KO] darwin/mips64le
[KO] darwin/mipsle
[KO] darwin/ppc64le
[KO] darwin/s390x
[KO] dragonfly/386
[KO] dragonfly/arm
[KO] dragonfly/arm64
[KO] dragonfly/mips
[KO] dragonfly/mips64
[KO] dragonfly/mips64le
[KO] dragonfly/mipsle
[KO] dragonfly/ppc64le
[KO] dragonfly/s390x
[KO] freebsd/arm64
[KO] freebsd/mips
[KO] freebsd/mips64
[KO] freebsd/mips64le
[KO] freebsd/mipsle
[KO] freebsd/ppc64le
[KO] freebsd/s390x
[KO] netbsd/arm64
[KO] netbsd/mips
[KO] netbsd/mips64
[KO] netbsd/mips64le
[KO] netbsd/mipsle
[KO] netbsd/ppc64le
[KO] netbsd/s390x
[KO] openbsd/arm64
[KO] openbsd/mips
[KO] openbsd/mips64
[KO] openbsd/mips64le
[KO] openbsd/mipsle
[KO] openbsd/ppc64le
[KO] openbsd/s390x
[KO] plan9/arm64
[KO] plan9/mips
[KO] plan9/mips64
[KO] plan9/mips64le
[KO] plan9/mipsle
[KO] plan9/ppc64le
[KO] plan9/s390x
[KO] solaris/386
[KO] solaris/arm
[KO] solaris/arm64
[KO] solaris/mips
[KO] solaris/mips64
[KO] solaris/mips64le
[KO] solaris/mipsle
[KO] solaris/ppc64le
[KO] solaris/s390x
[KO] windows/arm64
[KO] windows/mips
[KO] windows/mips64
[KO] windows/mips64le
[KO] windows/mipsle
[KO] windows/ppc64le
[KO] windows/s390x
```
