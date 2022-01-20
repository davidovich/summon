#!/bin/bash

echo
echo Hello {{ env "USER" }}
echo
echo Hello is run on the \"{{ now | date "02-01-2006" }}\" day
echo Hello args: \"$@\"
echo

{{if contains "env" (.osArgs | join "") }}
echo Following is the current environment:
echo
env
{{end}}
