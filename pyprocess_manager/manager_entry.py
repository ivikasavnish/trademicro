#!/usr/bin/env python3
"""
manager_entry.py: Starts dhanfeed_sync.py, updater.py, and trade_log.py as subprocesses, tracks them, and publishes status to Ably.
"""
import subprocess
import threading
import os
import signal
import sys
import time
import ably
import json

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
ABLY_CHANNEL = os.getenv("ABLY_CHANNEL", "process-control")

class ScriptProcess:
    def __init__(self, script, args):
        self.script = script
        self.args = args
        self.proc = None
        self.status = 'stopped'

    def start(self):
        if self.proc is None or self.proc.poll() is not None:
            cmd = ['python3', self.script] + self.args
            self.proc = subprocess.Popen(cmd)
            self.status = 'running'
        return self.proc.pid if self.proc else None

    def stop(self):
        if self.proc and self.proc.poll() is None:
            self.proc.terminate()
            self.status = 'stopped'
            return True
        return False

    def is_running(self):
        return self.proc and self.proc.poll() is None

class ProcessManager:
    def __init__(self):
        self.scripts = {
            'dhanfeed_sync.py': ScriptProcess('dhanfeed_sync.py', []),
            'updater.py': ScriptProcess('updater.py', []),
            'trade_log.py': ScriptProcess('trade_log.py', []),
        }
        self.ably = ably.AblyRest(ABLY_API_KEY)
        self.channel = self.ably.channels.get(ABLY_CHANNEL)
        self.listen_thread = threading.Thread(target=self.listen_commands, daemon=True)
        self.listen_thread.start()

    def listen_commands(self):
        for message in self.channel.subscribe():
            try:
                data = json.loads(message.data)
                action = data.get("action")
                script = data.get("script")
                args = data.get("args", [])
                if script not in self.scripts:
                    continue
                proc = self.scripts[script]
                if action == "start":
                    pid = proc.start()
                    self.publish_status('started', script, args, pid)
                elif action == "stop":
                    stopped = proc.stop()
                    self.publish_status('stopped', script, args, proc.proc.pid if stopped else None)
                elif action == "status":
                    self.publish_status(proc.status, script, args, proc.proc.pid if proc.proc else None)
            except Exception as e:
                print(f"Ably message error: {e}")

    def publish_status(self, status, script, args, pid=None, error=None):
        msg = {
            "status": status,
            "script": script,
            "args": args,
            "pid": pid,
            "error": error
        }
        self.channel.publish("status", json.dumps(msg))

if __name__ == "__main__":
    mgr = ProcessManager()
    print("Python manager_entry.py running. Listening for start/stop commands on Ably.")
    try:
        while True:
            time.sleep(5)
    except KeyboardInterrupt:
        print("Exiting manager.")
        sys.exit(0)
