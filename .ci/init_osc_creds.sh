#!/bin/bash
set -e

# this file is intended to be run inside a shap/continuous_deliver container
# see https://github.com/arbulu89/continuous-delivery

source /scripts/utils.sh

OSCRC_FILE=${OSCRC_FILE:=/root/.config/osc/oscrc}

check_user
