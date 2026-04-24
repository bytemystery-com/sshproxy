#!/bin/bash
# Copyright (c) 2026 Reiner Pröls
#
# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
# OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
# SOFTWARE.
#
# SPDX-License-Identifier: MIT

# CGO_ENABLED=1 CXX=x86_64-w64-mingw32-g++ CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" .
# CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" .
# CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC=$CC CXX=$CXX fyne package --os android --release --archs arm64,armv7 --metadata buildts="${ts}"

# sudo apt-get install libgl1-mesa-dev xorg-dev libxkbcommon-dev mingw-w64 osslsigncode cloc jq

# Zertifikat erzeugen für Windows
# openssl genrsa -out bytemystery_key.pem 4096
# openssl req -new -x509 -key bytemystery_key.pem -out bytemystery_cert.pem -days 10000 -subj "/C=DE/ST=Bayern/L=Munich/CN=bytemystery.com/emailAddress=bytemystery@gmail.com" -addext "subjectAltName=email:bytemystery@gmail.com"
# openssl pkcs12 -export -out bytemystery.pfx -inkey bytemystery_key.pem -in bytemystery_cert.pem

# Zertifikat für Android
# keytool -genkeypair -alias bytemystery -keyalg RSA -keysize 2048 -validity 10000 -keystore bytemystery.com.keystore -storetype PKCS12 -dname "CN=bytemystery.com, OU=IT, O=ByteMystery, L=Munich, ST=Bayern, C=DE, EMAILADDRESS=bytemystery@gmail.com"
# Anzeigen:
# apksigner verify --verbose MyApp.apk

PROGRAM_NAME=$(sed -n 's/^Name *= *"\(.*\)"/\1/p' FyneApp.toml)
VER=$(sed -n 's/^Version *= *"\(.*\)"/\1/p' FyneApp.toml)

BUILD=$(sed -n 's/^Build *= *\([0-9]\+\)/\1/p' FyneApp.toml)
VERSION=${VER}"."${BUILD}
NAME=${PROGRAM_NAME}
DESCRIPTION="SOCKS and HTTP Proxy over SSH connection"
COPYRIGHT="Reiner Pröls, bytemystery.com"
TRADEMARK="bytemystery.com"
COMPANY="bytemystery.com"
COMMENT="${DESCRIPTION}"

export PATH=${PATH}:~/go/bin
SED_TXT='sed.txt'

X=$(which osslsigncode)
if [[ ${X} == "" ]] ; then
    echo -e "osslsigncode must be installed\nsudo apt install osslsigncode"
    exit 1
fi

X=$(which jq)
if [[ ${X} == "" ]] ; then
    echo -e "jq must be installed\nsudo apt install jq"
    exit 1
fi

ONLYWIN=0
ONLYLINUX=0
ONLYAND=0
CLEAN=0
while getopts claw?h var
do
  case ${var} in
    l ) ONLYLINUX=1
	;;
    w ) ONLYWIN=1
	;;
    c ) CLEAN=1
	;;
    a ) ONLYAND=1
	;;
	* )
		usage
 		exit 1
	;;
  esac
done
shift $(($OPTIND - 1))

if [[ ${ONLYWIN} -eq 0 && ${ONLYAND} -eq 0 && ${ONLYLINUX} -eq 0 ]] ; then
    go fmt ./...    
    go vet ./... 
    go get -u ./... # ???   
    go mod tidy
    if [[ ${CLEAN} ]] ; then
        go clean -cache -modcache
    fi
    ONLYWIN=1
    ONLYLINUX=1
    ONLYAND=1
fi

go install fyne.io/tools/cmd/fyne@latest
if [[ ${ONLYWIN} -eq 1 ]] ; then
    go get github.com/tc-hib/go-winres@latest
    go install github.com/tc-hib/go-winres@latest
fi

ts=$(date -u +'%Y-%m-%d - %H:%M:%S')
rm assets/lang/xx.json
fyne translate assets/lang/xx.json
cp assets/lang/xx.json assets/lang/en.json
LANGUAGES="de"
for la in ${LANGUAGES} ; do
    jq -s '.[0] * .[1]' assets/lang/xx.json assets/lang/${la}.json > assets/lang/${la}_x.json
    mv assets/lang/${la}_x.json assets/lang/${la}.json
done

cloc .
TAGS="NOTAG"
for tag in ${TAGS} ; do
    echo "Build tag: ${tag} ..."
    if [[ ${tag} == "NOTAG" ]] ; then
        suffix=""
    else 
        suffix="_${tag}"
    fi
    if [[ ${ONLYWIN} -eq 1 ]] ; then
        export CC_WIN=x86_64-w64-mingw32-gcc
        export CXX_WIN=x86_64-w64-mingw32-g++
        echo "Build windows ..."
        CGO_ENABLED=1 CXX=${CXX_WIN} CC=${CC_WIN} GOOS=windows GOARCH=amd64 fyne package --release --metadata buildts="${ts}" --tags ${tag}
        # GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" .
        rm "${SED_TXT}" 2>/dev/null
        echo "s/<VERSION>/${VERSION}/g" >>"${SED_TXT}"
        echo "s/<DESCRIPTION>/${DESCRIPTION}/g" >>"${SED_TXT}"
        echo "s/<BUILD>/${BUILD}/g" >>"${SED_TXT}"
        echo "s/<NAME>/${NAME}/g" >>"${SED_TXT}"
        echo "s/<COPYRIGHT>/${COPYRIGHT}/g" >>"${SED_TXT}"
        echo "s/<TRADEMARK>/${TRADEMARK}/g" >>"${SED_TXT}"
        echo "s/<COMPANY>/${COMPANY}/g" >>"${SED_TXT}"
        echo "s/<COMMENT>/${COMMENT}/g" >>"${SED_TXT}"
        sed -f "${SED_TXT}" winres.json > winres_act.json
        rm "${SED_TXT}"
        # go-winres make --in winres_act.json
        go-winres patch --in winres_act.json --no-backup --delete "${PROGRAM_NAME}".exe
        rm winres_act.json
        WINDOWS_KEY=~/Dropbox/private/windows_key/windows_key.txt
        if [[ -f "${WINDOWS_KEY}" ]] ; then
            dir=$(dirname "${WINDOWS_KEY}")
            WIN_PFX="${dir}"/$(tail -n+1 "${WINDOWS_KEY}" | head -n1)
            WIN_PASS=$(tail -n+2 "${WINDOWS_KEY}" | head -n1)
            osslsigncode sign -pkcs12 "${WIN_PFX}" -pass "${WIN_PASS}" -n ""${PROGRAM_NAME}"" -ts http://timestamp.digicert.com -in "${PROGRAM_NAME}".exe -out "${PROGRAM_NAME}"_sig.exe
            rm "${PROGRAM_NAME}".exe
            mv "${PROGRAM_NAME}"_sig.exe "${PROGRAM_NAME}"${suffix}.exe
        fi
        mkdir -p dist/windows
        mv "${PROGRAM_NAME}"${suffix}.exe dist/windows
    fi

    if [[ ${ONLYLINUX} -eq 1 ]] ; then
        echo "Build linux ..."
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 fyne package --release --executable "${PROGRAM_NAME}" --metadata buildts="${ts}" --tags ${tag}
        CGO_ENABLED=1 GOOS=linux GOARCH=amd64 fyne build --release --output "${PROGRAM_NAME}" --metadata buildts="${ts}" --tags ${tag}
        # go build -ldflags="-s -w" .
        mkdir -p dist/linux
        sudo mv -f "${PROGRAM_NAME}" dist/linux/"${PROGRAM_NAME}"${suffix}
        mv "${PROGRAM_NAME}".tar.xz dist/linux/"${PROGRAM_NAME}"${suffix}.tar.xz
    fi

    if [[ ${ONLYAND} -eq 1 ]] ; then
        # Android stuff
        echo "Build android ..."
        ANDROID_KEY=~/Dropbox/private/android_key/android_key.txt
        HAS_ANDROID_KEY=0
        if [[ -f "${ANDROID_KEY}" ]] ; then
            dir=$(dirname "${ANDROID_KEY}")
            ANDROID_KEY_FILE="${dir}"/$(tail -n+1 "${ANDROID_KEY}" | head -n1)
            ANDROID_KEYSTORE_PASS=$(tail -n+2 "${ANDROID_KEY}" | head -n1)
            ANDROID_KEY_PASS=$(tail -n+3 "${ANDROID_KEY}" | head -n1)
            ANDROID_KEY_ALIAS=$(tail -n+4 "${ANDROID_KEY}" | head -n1)
            HAS_ANDROID_KEY=1
        fi
        mkdir -p dist/android
        export ANDROID_HOME=${HOME}/Android
        export ANDROID_NDK_HOME=${ANDROID_HOME}/ndk/latest
        OLD_PATH=${PATH}
        export PATH=$PATH:$ANDROID_HOME/platform-tools:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/build-tools/latest
        fyne package -os android --release --metadata buildts="${ts}" --tags ${tag}
        if [[ ${HAS_ANDROID_KEY} -eq 1 ]] ; then
            apksigner sign --ks "${ANDROID_KEY_FILE}" --ks-key-alias "${ANDROID_KEY_ALIAS}" --ks-pass "pass:${ANDROID_KEYSTORE_PASS}" --key-pass "pass:${ANDROID_KEY_PASS}" "${PROGRAM_NAME}".apk
            rm "${PROGRAM_NAME}".apk.idsig
            mv "${PROGRAM_NAME}".apk dist/android/"${PROGRAM_NAME}"${suffix}.apk
        fi
        fyne package -target android/arm64 --release --metadata buildts="${ts}" --tags ${tag}
        if [[ ${HAS_ANDROID_KEY} -eq 1 ]] ; then
            apksigner sign --ks "${ANDROID_KEY_FILE}" --ks-key-alias "${ANDROID_KEY_ALIAS}" --ks-pass "pass:${ANDROID_KEYSTORE_PASS}" --key-pass "pass:${ANDROID_KEY_PASS}" "${PROGRAM_NAME}".apk
            rm "${PROGRAM_NAME}".apk.idsig
            mv "${PROGRAM_NAME}".apk dist/android/"${PROGRAM_NAME}"${suffix}_64.apk
        fi
        if [[ ${HAS_ANDROID_KEY} -eq 1 ]] ; then
            fyne release --target android/arm --keystore "${ANDROID_KEY_FILE}" --keystore-pass "${ANDROID_KEYSTORE_PASS}" --key-pass "${ANDROID_KEY_PASS}" --key-name "${ANDROID_KEY_ALIAS}" --metadata buildts="${ts}" --tags ${tag}
            mv "${PROGRAM_NAME}".aab dist/android/"${PROGRAM_NAME}"${suffix}_32.aab
            fyne release --target android/arm64 --keystore "${ANDROID_KEY_FILE}" --keystore-pass "${ANDROID_KEYSTORE_PASS}" --key-pass "${ANDROID_KEY_PASS}" --key-name "${ANDROID_KEY_ALIAS}" --metadata buildts="${ts}" --tags ${tag}
            mv "${PROGRAM_NAME}".aab dist/android/"${PROGRAM_NAME}"${suffix}_64.aab
        fi
        PATH=${OLD_PATH}
    fi
done

