import os
import asyncio
import json
from ably import AblyRealtime

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

if not ABLY_API_KEY:
    raise ValueError("Set ABLY_API_KEY as an environment variable or in this script.")

async def main():
    ably = AblyRealtime(ABLY_API_KEY)
    channel = ably.channels.get(CHANNEL_NAME)
    print(f"Listening for all messages on channel '{CHANNEL_NAME}'...")

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

    await channel.subscribe(on_message)
    while True:
        await asyncio.sleep(3600)

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nExiting Ably listener.")
