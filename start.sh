#!/bin/sh

echo Installing plugins...

chmod +x ./plugins/install.sh
./plugins/install.sh

echo Starting ControlHub...
eval $@
