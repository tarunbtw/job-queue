#!/bin/sh
set -e

# Start the worker in the background.
# If the worker crashes, it stays down; the container keeps running
# because the server (foreground) is the health signal.
./worker &

# Run the server in the foreground.
# The platform will restart the container if the server exits.
exec ./server
