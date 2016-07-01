#!/bin/bash

function handleline() {
	local name nargs inthing args resp
	case "$1" in
	Comm) inthing=""; ;;
	Room) inthing=" (in room)"; ;;
	Game) inthing=" (in game)"; ;;
	esac
	shift

	name=$1; shift
	nargs=$1; shift
	args=$1; shift
	resp="$*"

	printf "%s %s%s -> %s\n" "$name" "$args" "$inthing" "$resp"
}

lines="$(
	grep 'handle\(Room\|Game\)\?Command(' wshandler.go \
	| sed 's/\t\+//' \
	| grep '//HC' \
	| tr -d \" \
	| sed 's/,,\? / /g' \
	| sed 's/ func() { \/\/HC / /' \
	| sed 's/handleCommand/Comm/' \
	| tr \( \  \
	| sed 's/handle\|Command//g')"
IFS=$'\n'
for line in $lines; do
	IFS=" "
	handleline $line
done
