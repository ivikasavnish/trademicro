
import os
import asyncio
import json
from ably import AblyRest, AblyRealtime
"""
AblyProcessManager: Listen to Ably messages and execute/manage subprocesses.

Supported commands:
- start: Launch a process (python3 dhanfeed_sync.py, updater.py, or trade_log.py)
- stop:  Stop a running process
- resume: (Optional) Resume a stopped process if supported
- restart: Restart a process (stop if running, then start)
- status: Query the state of all managed processes

Advanced features:
- Error reporting: Any errors (e.g., failed start/stop) are published as event "error" with details.
- Process output capture: All stdout/stderr from managed processes is published as event "output" with the script name and output.

Ably message/event structure:
- Event name: "command" (recommended)
- Payload (JSON):
    {
        "action": "start" | "stop" | "resume" | "status",
        "script": "dhanfeed_sync.py" | "updater.py" | "trade_log.py"   # Not needed for status
    }

Example (to start dhanfeed_sync.py):
    Event: "command"
    Payload: { "action": "start", "script": "dhanfeed_sync.py" }

Example (to query status):
    Event: "command"
    Payload: { "action": "status" }

Response events:
- Event: "state"
- Payload: { "script": <script>, "status": "running"|"stopped"|"not_started", "pid": <pid or null> }
- For status query, sends state for all managed scripts.

"""
# Load environment variables from .env if present 
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass  # If python-dotenv is not installed, ignore

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

import subprocess

class AblyProcessManager:
    MANAGED_SCRIPTS = ["dhanfeed_sync.py", "updater.py", "trade_log.py"]

    def __init__(self, api_key=None, channel_name=None):
        self.api_key = api_key or ABLY_API_KEY
        self.channel_name = channel_name or CHANNEL_NAME
        if not self.api_key:
            raise ValueError("ABLY_API_KEY must be set as an environment variable or passed explicitly.")
        self.rest = AblyRest(self.api_key)
        self.realtime = AblyRealtime(self.api_key)
        self.rest_channel = self.rest.channels.get(self.channel_name)
        self.rt_channel = self.realtime.channels.get(self.channel_name)
        self._callback = None
        # Track subprocesses: {script_name: subprocess.Popen}
        self.processes = {}
        # Track state: {script_name: "running"|"stopped"|"not_started"}
        self.state = {script: "not_started" for script in self.MANAGED_SCRIPTS}

    def set_message_callback(self, callback):
        """Register a callback to be called with (name, data) for each incoming message."""
        self._callback = callback

    async def listen(self):
        async def on_message(message):
            data = message.data
            try:
                payload = json.loads(data)
            except Exception:
                payload = data
            # Process Ably command events
            await self.handle_event(message.name, payload)
            # User callback (optional)
            if self._callback:
                await self._callback(message.name, payload)
            else:
                print(f"[Ably] Received message: {message.name}: {payload}")
        await self.rt_channel.subscribe(on_message)
        # Keep the listener alive indefinitely
        while True:
            await asyncio.sleep(3600)

    def publish(self, name, data):
        if not isinstance(data, str):
            data = json.dumps(data)
        self.rest_channel.publish(name, data)

    async def handle_event(self, event_name, payload):
        # Only handle 'command' events for process control
        if event_name != "command":
            return
        if not isinstance(payload, dict):
            self.publish_error("Invalid payload", payload)
            return
        action = payload.get("action")
        script = payload.get("script")
        if action == "start" and script in self.MANAGED_SCRIPTS:
            self.start_process(script)
            self.publish_state(script)
        elif action == "stop" and script in self.MANAGED_SCRIPTS:
            self.stop_process(script)
            self.publish_state(script)
        elif action == "resume" and script in self.MANAGED_SCRIPTS:
            if self.state[script] != "running":
                self.start_process(script)
            self.publish_state(script)
        elif action == "restart" and script in self.MANAGED_SCRIPTS:
            self.restart_process(script)
            self.publish_state(script)
        elif action == "status":
            self.publish_all_states()
        else:
            self.publish_error(f"Unknown command or script: {payload}", payload)

    def start_process(self, script):
        if self.state[script] == "running":
            self.publish_error(f"{script} is already running.")
            return
        try:
            proc = subprocess.Popen(["python3", script], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            self.processes[script] = proc
            self.state[script] = "running"
            # Start output capture
            self.capture_output(script, proc)
        except Exception as e:
            self.publish_error(f"Failed to start {script}: {e}")
            self.state[script] = "not_started"

    def stop_process(self, script):
        proc = self.processes.get(script)
        if proc and self.state[script] == "running":
            try:
                proc.terminate()
                proc.wait(timeout=5)
                self.state[script] = "stopped"
            except Exception as e:
                self.publish_error(f"Failed to stop {script}: {e}")
        else:
            self.publish_error(f"{script} is not running.")

    def publish_state(self, script):
        proc = self.processes.get(script)
        state = {
            "script": script,
            "status": self.state[script],
            "pid": proc.pid if proc and self.state[script] == "running" else None
        }
        asyncio.create_task(self.publish("state", state))

    def restart_process(self, script):
        self.stop_process(script)
        self.start_process(script)

    def publish_error(self, message, payload=None):
        error = {"error": message}
        if payload is not None:
            error["payload"] = payload
        asyncio.create_task(self.publish("error", error))

    def capture_output(self, script, proc):
        import threading
        def read_stream(stream, stream_type):
            for line in iter(stream.readline, ''):
                asyncio.create_task(self.publish("output", {"script": script, "stream": stream_type, "output": line.rstrip()}))
            stream.close()
        threading.Thread(target=read_stream, args=(proc.stdout, "stdout"), daemon=True).start()
        threading.Thread(target=read_stream, args=(proc.stderr, "stderr"), daemon=True).start()

    def publish_all_states(self):
        for script in self.MANAGED_SCRIPTS:
            self.publish_state(script)

    def set_message_callback(self, callback):
        """Register a callback to be called with (name, data) for each incoming message."""
        self._callback = callback

    async def listen(self):
        async def on_message(message):
            data = message.data
            try:
                payload = json.loads(data)
            except Exception:
                payload = data
            if self._callback:
                await self._callback(message.name, payload)
            else:
                print(f"[Ably] Received message: {message.name}: {payload}")
        await self.rt_channel.subscribe(on_message)
        # Keep the listener alive indefinitely
        while True:
            await asyncio.sleep(3600)

    def publish(self, name, data):
        if not isinstance(data, str):
            data = json.dumps(data)
        self.rest_channel.publish(name, data)

if __name__ == "__main__":
    import sys
    async def main():
        print(f"Connecting to Ably channel '{CHANNEL_NAME}'...")
        mgr = AblyProcessManager()
        async def print_callback(name, data):
            print(f"[Ably] Message received: {name}: {data}")
        mgr.set_message_callback(print_callback)
        print("Listening for messages. Press Ctrl+C to exit.")
        await mgr.listen()
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nExiting Ably manager.")
