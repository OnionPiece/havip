#!/bin/bash

ABS_PATH="$(cd `dirname $0`; pwd)"
bash $ABS_PATH/utils/api_main.sh $@
