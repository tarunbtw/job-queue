#!/bin/sh
set -e

(sleep 2 && ./worker) &
exec ./server