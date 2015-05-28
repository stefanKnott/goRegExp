#/!/bin/bash

#TIMEFORMAT="wall=%e user=%U system=%S CPU=%P"
FILENAME="input.txt"

echo Building code...
go build project.go
time ./project $FILENAME