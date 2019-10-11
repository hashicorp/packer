rm -rf crosstest
mkdir -p crosstest

export GOOS=linux;   export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=linux;   export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=darwin;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=darwin;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=freebsd; export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=freebsd; export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=openbsd; export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=openbsd; export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=netbsd;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=netbsd;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=dragonfly;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=dragonfly;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=solaris;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=solaris;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=plan9;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=plan9;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=windows;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=windows;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";

export GOOS=android;  export GOARCH=arm;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=arm64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=amd64;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=386;       echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=mips;      echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=mipsle;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=mips64;    echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=mips64le;  echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=ppc64le;   echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
export GOOS=android;  export GOARCH=s390x;     echo "$(go build -o crosstest/${GOOS}_${GOARCH} && echo '[OK]' || echo '[KO]') $GOOS/$GOARCH";
