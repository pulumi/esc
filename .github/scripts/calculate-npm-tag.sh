#!/bin/bash
if [[ ! "$1" =~ "-" ]]; then
  echo "latest"
elif [[ "$1" =~ "-beta" ]]; then
  echo "beta"
elif [[ "$1" =~ "-rc" ]]; then
  echo "rc"
else
  echo "dev"
fi