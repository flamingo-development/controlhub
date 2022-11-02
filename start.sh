#!/bin/sh

echo Installing plugins...
./plugins/install.sh

echo Starting ControlHub...
eval $@
