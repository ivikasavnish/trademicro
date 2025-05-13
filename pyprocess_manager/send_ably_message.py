import os
import json
import asyncio
from ably_manager import AblyProcessManager

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

if not ABLY_API_KEY:
    raise ValueError("Set ABLY_API_KEY as an environment variable or in this script.")

async def listen_for_messages(mgr):
    async def print_callback(name, data):
        print(f"[Ably] Message received: {name}: {data}")
    mgr.set_message_callback(print_callback)
    print("Listening for messages. Press Ctrl+C to stop.")
    try:
        await mgr.listen()
    except KeyboardInterrupt:
        print("\nStopped listening.")

async def main():
    mgr = AblyProcessManager(api_key=ABLY_API_KEY, channel_name=CHANNEL_NAME)
    while True:
        print("\nAbly Interactive Test Menu:")
        print("1. Publish a message")
        print("2. Listen for messages")
        print("3. Exit")
        choice = input("Select an option (1-3): ").strip()
        if choice == '1':
            event = input("Enter event name: ").strip()
            data = input("Enter message data (JSON or string): ").strip()
            try:
                parsed_data = json.loads(data)
            except Exception:
                parsed_data = data
            mgr.publish(event, parsed_data)
            print(f"Message published to '{CHANNEL_NAME}' as event '{event}': {parsed_data}")
        elif choice == '2':
            await listen_for_messages(mgr)
        elif choice == '3':
            print("Exiting.")
            break
        else:
            print("Invalid option. Please choose 1, 2, or 3.")

if __name__ == "__main__":
    asyncio.run(main())
