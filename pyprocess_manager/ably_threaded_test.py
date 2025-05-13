import os
import json
import threading
import time
import asyncio
from ably import AblyRealtime

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

if not ABLY_API_KEY:
    raise ValueError("Set ABLY_API_KEY as an environment variable or in this script.")

# Thread-safe sender
class AblySender:
    def __init__(self, ably, channel):
        self.ably = ably
        self.channel = channel
    def send(self, event_name, data):
        if not isinstance(data, str):
            data = json.dumps(data)
        self.channel.publish(event_name, data)
        print(f"[SENT] Event: {event_name} | Data: {data}")

# Threaded receiver
class AblyReceiver(threading.Thread):
    def __init__(self, ably, channel):
        super().__init__(daemon=True)
        self.ably = ably
        self.channel = channel
    def run(self):
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        loop.run_until_complete(self.listen())
    async def listen(self):
        async def on_message(message):
            try:
                data = message.data
                try:
                    data = json.loads(data)
                except Exception:
                    pass
                print(f"[RECEIVED] Event: {message.name} | Data: {data}")
            except Exception as e:
                print(f"[ERROR] Failed to parse message: {e}")
        await self.channel.subscribe(on_message)
        while True:
            await asyncio.sleep(3600)

def main():
    ably = AblyRealtime(ABLY_API_KEY)
    channel = ably.channels.get(CHANNEL_NAME)
    sender = AblySender(ably, channel)
    receiver = AblyReceiver(ably, channel)
    receiver.start()

    # Simple interactive loop for sending
    while True:
        print("\nAbly Threaded Test Menu:")
        print("1. Send a command event")
        print("2. Send a status query")
        print("3. Exit")
        choice = input("Select an option (1-3): ").strip()
        if choice == '1':
            script = input("Script name (e.g. dhanfeed_sync.py): ").strip()
            action = input("Action (start/stop/restart): ").strip()
            payload = {"action": action, "script": script}
            sender.send("command", payload)
        elif choice == '2':
            sender.send("command", {"action": "status"})
        elif choice == '3':
            print("Exiting.")
            break
        else:
            print("Invalid option.")
        time.sleep(0.5)

if __name__ == "__main__":
    main()
