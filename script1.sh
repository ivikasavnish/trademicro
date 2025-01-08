#!/bin/bash

SCRIPT_PATH="updater.py"
PID_FILE="updater.pid"

# Make script executable
chmod +x "$SCRIPT_PATH"
# source .venv/bin/activate
# Run with nohup, redirect output to /dev/null
nohup /root/trademicro/trademicro/.venv/bin/python3 "$SCRIPT_PATH" > /dev/null 2>&1 &

# Save PID
echo $! > "$PID_FILE"

echo "Started dhanfeed_sync.py with PID $(cat "$PID_FILE")"
echo "To check status: ps -p $(cat "$PID_FILE")"
echo "To stop: kill $(cat "$PID_FILE")"