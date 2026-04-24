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
#
# Author: Reiner Pröls

# Create Host key
# ssh-keygen -t ed25519 -f <XXX> -C "sshproxy host key" -N ""
# xxx is the private key - this is what the SSH server needs
# XXX.pub is the public key which is needed by sshproxy
# Create User key
# ssh-keygen -t ed25519 -f <YYY> -C "<user@mail.com>" -N "<PASS>"
# yyy.pub is the public key - this is what the SSH server needs in authorized_keys
# YYY is the public key which is needed by sshproxy
# PASS must be given in sshproxy


HOSTKEY="sshproxy_host"
KEYFILE="sshproxy_user"
AUTHKEYFILE="auth_key"
PORT=8222

CFGFILE="sshd_sshproxy"
CFGFILETEMP="sshd_temp"
SSHDOPTIONS=("-D -e")

SSHD=$(which sshd)

function check_progs()
{
	local ret=0
	local progs="${SSHD} ssh-keygen envsubst"
	local p
	for p in ${progs} ; do
		if ! command -v "${p}" >/dev/null 2>&1; then
			echo "!!! Missing program '${p}' - please install it !!!" 
			ret=1
		fi
	done
	return ${ret}
}

function create_hostkey()
{
	ssh-keygen -t ed25519 -o -f "${HOSTKEY}" -C "sshproxy ${HOSTNAME} - $(date +%Y-%m-%d)" -N ""
	chmod 600 "${HOSTKEY}"
	chmod 600 "${HOSTKEY}.pub"
}

function create_userkey()
{
	ssh-keygen -t ed25519 -a 100 -o -f "${KEYFILE}" -C "${1}" -N "${2}"
	cat "${KEYFILE}.pub" > "${AUTHKEYFILE}"
	chmod 600 "${KEYFILE}"
	chmod 600 "${KEYFILE}.pub"
	chmod 600 "${AUTHKEYFILE}"
}

function prepare_cfg()
{
	HOSTKEY=$(readlink -f "${HOSTKEY}")
	export HOSTKEY=${HOSTKEY}
	chmod 600 "${HOSTKEY}"
	
	AUTHKEYFILE=$(readlink -f "${AUTHKEYFILE}")
	chmod 600 "${AUTHKEYFILE}"
	export AUTHKEYFILE=${AUTHKEYFILE}
	
	export PORT=${PORT}
	envsubst < "${CFGFILE}" > "${CFGFILETEMP}"
}

function stop_sshd()
{
	if [[ ${PID} != "" ]] ; then
		kill -9 ${PID}
	fi
	rm "${CFGFILETEMP}"
}

# Trap CTRL+C
function do_exit ()
{
    echo -e "\nCTRL+C pressed. Cleanup."
	stop_sshd
    exit 0
}

function ssh_msg()
{
	clear
	echo "Using port '${PORT}' as user '${USER}'"
	echo ""
	echo "Press CTRL+C for terminating"
}

function usage()
{
	echo "Run ./start.sh -c USER PASS"
	echo "   for creating hostkey and user key files"
	echo "Run ./start.sh"
	echo "   for starting the SSH server"
}

CREATEKEYS=0
while getopts c?h var
do
  case ${var} in
    c ) CREATEKEYS=1
	;;
	* )
		usage
 		exit 1
	;;
  esac
done
shift $(($OPTIND - 1))

check_progs
if [[ $? -ne 0 ]] ; then
	exit 1
fi

cd "$(dirname $0)"

if [[ ${CREATEKEYS} -eq 1 ]] ; then
	create_hostkey
	create_userkey "${@}"
else 
	trap do_exit SIGINT
	ssh_msg
	prepare_cfg
	"${SSHD}" ${SSHDOPTIONS[@]} -f ${CFGFILETEMP} &
	PID=$!
	while true; do
    	sleep 1
	done	
fi
