# This gRPC server also supports Ably Pub/Sub for process control.
# On startup, it launches AblyProcessManager to listen for start/stop/resume commands via Ably channels.

import time
from ably_manager import AblyProcessManager

if __name__ == '__main__':
    print("Process Manager (Ably-only mode) starting...")
    # Dummy process manager for AblyProcessManager
    class DummyManager:
        def StartProcess(self, req, ctx):
            print(f"[Ably] Received start: script={getattr(req, 'script', None)}, args={getattr(req, 'args', None)}")
        def StopProcess(self, req, ctx):
            print(f"[Ably] Received stop: script={getattr(req, 'script', None)}, args={getattr(req, 'args', None)}")
        def ResumeProcess(self, req, ctx):
            print(f"[Ably] Received resume: script={getattr(req, 'script', None)}, args={getattr(req, 'args', None)}")
    AblyProcessManager(DummyManager())
    print("Listening for Ably messages...")
    try:
        while True:
            time.sleep(86400)
    except KeyboardInterrupt:
        print("Exiting Ably-only server.")
