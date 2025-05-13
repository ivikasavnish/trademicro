import os
import asyncio
import json
from ably import AblyRest, AblyRealtime
import subprocess

"""
CustomAblyProcessManager: Listen to Ably messages and execute/manage only selected subprocesses.

This version is tailored for a specific subset of scripts you want to control, without touching the generic AblyProcessManager.

To use: set MANAGED_SCRIPTS below to the exact scripts you want to allow.

Supported commands:
- start: Launch a process (python3 <selected_script>)
- stop:  Stop a running process
- restart: Restart a process (stop if running, then start)
- status: Query the state of all managed processes

Advanced features:
- Error reporting: Any errors (e.g., failed start/stop) are published as event "error" with details.
- Process output capture: All stdout/stderr from managed processes is published as event "output" with the script name and output.

Ably message/event structure:
- Event name: "command"
- Payload (JSON):
    {
        "action": "start" | "stop" | "restart" | "status",
        "script": <your_script.py>   # Not needed for status
    }

Example (to start a script):
    Event: "command"
    Payload: { "action": "start", "script": "my_script.py" }

Example (to query status):
    Event: "command"
    Payload: { "action": "status" }
"""

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

class CustomAblyProcessManager:
    # CHANGE THIS LIST to the scripts you want to allow
    MANAGED_SCRIPTS = ["dhanfeed_sync.py", "updater.py"]  # Example: only allow these two

    # Map ably event names to process actions
    EVENT_ACTION_MAP = {
        "feed_start":   {"action": "start", "script": "dhanfeed_sync.py"},
        "feed_stop":    {"action": "stop",  "script": "dhanfeed_sync.py"},
        "updater":      {"action": "start", "script": "updater.py"},
        "updater_stop": {"action": "stop",  "script": "updater.py"},
        # Add more mappings as needed
    }

    def __init__(self, api_key=None, channel_name=None):
        self.api_key = api_key or ABLY_API_KEY
        self.channel_name = channel_name or CHANNEL_NAME
        if not self.api_key:
            raise ValueError("ABLY_API_KEY must be set as an environment variable or passed explicitly.")
        self.rest = AblyRest(self.api_key)
        self.realtime = AblyRealtime(self.api_key)
        self.rest_channel = self.rest.channels.get(self.channel_name)
        self.rt_channel = self.realtime.channels.get(self.channel_name)
        self.processes = {}
        self.state = {script: "not_started" for script in self.MANAGED_SCRIPTS}

    async def listen(self):
        async def on_message(message):
            data = message.data
            print(f"[MANAGER] Received event: {message.name} | Raw data: {data}")
            try:
                payload = json.loads(data)
            except Exception:
                payload = data
            await self.handle_event(message.name, payload)
        await self.rt_channel.subscribe(on_message)
        print(f"[MANAGER] Subscribed to Ably channel '{self.channel_name}' and listening for events...")
        while True:
            await asyncio.sleep(3600)

    async def publish(self, name, data):
        if not isinstance(data, str):
            data = json.dumps(data)
        await self.rest_channel.publish(name, data)

    async def handle_event(self, event_name, payload):
        print(f"[MANAGER] Handling event: {event_name} | Payload: {payload}")
        # 1. Check if event_name is in the custom mapping
        if event_name in self.EVENT_ACTION_MAP:
            mapped = self.EVENT_ACTION_MAP[event_name]
            print(f"[MANAGER] Matched mapped event: {event_name} -> {mapped}")
            action = mapped.get("action")
            script = mapped.get("script")
            # Call the same logic as below
            if action == "start" and script in self.MANAGED_SCRIPTS:
                print(f"[MANAGER] Starting process: {script}")
                self.start_process(script)
                self.publish_state(script)
            elif action == "stop" and script in self.MANAGED_SCRIPTS:
                print(f"[MANAGER] Stopping process: {script}")
                self.stop_process(script)
                self.publish_state(script)
            elif action == "restart" and script in self.MANAGED_SCRIPTS:
                print(f"[MANAGER] Restarting process: {script}")
                self.restart_process(script)
                self.publish_state(script)
            elif action == "status":
                print(f"[MANAGER] Publishing all states.")
                self.publish_all_states()
            else:
                print(f"[MANAGER] Unknown mapped command or script: {mapped}")
                self.publish_error(f"Unknown mapped command or script: {mapped}", mapped)
            return
        # 2. Fallback to original command event logic
        if event_name != "command":
            print(f"[MANAGER] Ignoring event: {event_name}")
            return
        if not isinstance(payload, dict):
            print(f"[MANAGER] Invalid payload type: {type(payload)}")
            self.publish_error("Invalid payload", payload)
            return
        action = payload.get("action")
        script = payload.get("script")
        if action == "start" and script in self.MANAGED_SCRIPTS:
            print(f"[MANAGER] Starting process: {script}")
            self.start_process(script)
            self.publish_state(script)
        elif action == "stop" and script in self.MANAGED_SCRIPTS:
            print(f"[MANAGER] Stopping process: {script}")
            self.stop_process(script)
            self.publish_state(script)
        elif action == "restart" and script in self.MANAGED_SCRIPTS:
            print(f"[MANAGER] Restarting process: {script}")
            self.restart_process(script)
            self.publish_state(script)
        elif action == "status":
            print(f"[MANAGER] Publishing all states.")
            self.publish_all_states()
        else:
            print(f"[MANAGER] Unknown command or script: {payload}")
            self.publish_error(f"Unknown command or script: {payload}", payload)

    def start_process(self, script):
        if self.state[script] == "running":
            print(f"[MANAGER] {script} is already running.")
            self.publish_error(f"{script} is already running.")
            return
        try:
            print(f"[MANAGER] Launching subprocess: python3 {script}")
            proc = subprocess.Popen(["python3", script], stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            self.processes[script] = proc
            self.state[script] = "running"
            self.capture_output(script, proc)
        except Exception as e:
            print(f"[MANAGER] Failed to start {script}: {e}")
            self.publish_error(f"Failed to start {script}: {e}")
            self.state[script] = "not_started"

    def stop_process(self, script):
        proc = self.processes.get(script)
        if proc and self.state[script] == "running":
            try:
                print(f"[MANAGER] Terminating subprocess for {script}")
                proc.terminate()
                proc.wait(timeout=5)
                self.state[script] = "stopped"
            except Exception as e:
                print(f"[MANAGER] Failed to stop {script}: {e}")
                self.publish_error(f"Failed to stop {script}: {e}")
        else:
            print(f"[MANAGER] {script} is not running.")
            self.publish_error(f"{script} is not running.")

    def restart_process(self, script):
        print(f"[MANAGER] Restarting subprocess for {script}")
        self.stop_process(script)
        self.start_process(script)

    def publish_state(self, script):
        proc = self.processes.get(script)
        state = {
            "script": script,
            "status": self.state[script],
            "pid": proc.pid if proc and self.state[script] == "running" else None
        }
        print(f"[MANAGER] Publishing state: {state}")
        asyncio.create_task(self.publish("state", state))

    def publish_all_states(self):
        print(f"[MANAGER] Publishing all states for managed scripts...")
        for script in self.MANAGED_SCRIPTS:
            self.publish_state(script)

    def publish_error(self, message, payload=None):
        error = {"error": message}
        if payload is not None:
            error["payload"] = payload
        print(f"[MANAGER] Publishing error: {error}")
        asyncio.create_task(self.publish("error", error))

    def capture_output(self, script, proc):
        import threading
        def read_stream(stream, stream_type):
            for line in iter(stream.readline, ''):
                print(f"[MANAGER] Captured {stream_type} from {script}: {line.rstrip()}")
                asyncio.create_task(self.publish("output", {"script": script, "stream": stream_type, "output": line.rstrip()}))
            stream.close()
        threading.Thread(target=read_stream, args=(proc.stdout, "stdout"), daemon=True).start()
        threading.Thread(target=read_stream, args=(proc.stderr, "stderr"), daemon=True).start()

if __name__ == "__main__":
    async def main():
        mgr = CustomAblyProcessManager()
        print(f"Listening for messages on channel '{CHANNEL_NAME}' for scripts: {mgr.MANAGED_SCRIPTS}")
        await mgr.listen()
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nExiting Custom Ably manager.")
