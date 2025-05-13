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

    events = [
        ("feed_start", None),
        ("updater", None),
        # ("feed_stop", None),
        # ("updater_stop", None),
    ]
    delays = [0, 20]  # seconds between events

    for (event_name, data), delay in zip(events, delays):
        await asyncio.sleep(delay)
        await channel.publish(event_name, "" if data is None else data)
        print(f"[SENT] Event: {event_name}")

    print("All custom events sent.")

if __name__ == "__main__":
    asyncio.run(main())
