import os
import json
import asyncio
from ably import AblyRealtime

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

if not ABLY_API_KEY:
    raise ValueError("Set ABLY_API_KEY as an environment variable or in this script.")

async def main():
    ably = AblyRealtime(ABLY_API_KEY)
    channel = ably.channels.get(CHANNEL_NAME)
    msg = {
        "action": "start",
        "script": "trade_log.py",
        "args": []
    }
    await channel.publish("command", json.dumps(msg))
    print(f"Ably message sent to channel '{CHANNEL_NAME}': {msg}")

if __name__ == "__main__":
    asyncio.run(main())
