#!/bin/bash

function handleline() {
	local name nargs ingame args plural resp
	name=$1; shift
	nargs=$1; shift
	if test "$1" = "true"; then ingame=" (in game)"; else ingame=""; fi; shift
	args=$1; shift
	resp="$*"

	if test "$nargs" -eq 1; then plural=""; else plural="s"; fi

	printf "%s %s%s -> %s\n" "$name" "$args" "$ingame" "$resp"
}

lines="$(grep 'handleCommand(' wshandler.go | cut -d\( -f2- | tr -d \" | sed 's/,,\? / /g' | sed 's/ func() { \/\/HC / /')"
IFS=$'\n'
for line in $lines; do
	IFS=" "
	handleline $line
done
