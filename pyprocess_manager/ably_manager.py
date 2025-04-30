

import os
import asyncio
import json
from ably import AblyRest, AblyRealtime

# Load environment variables from .env if present
try:
    from dotenv import load_dotenv
    load_dotenv()
except ImportError:
    pass  # If python-dotenv is not installed, ignore

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

class AblyProcessManager:
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
