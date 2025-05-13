import os
import asyncio
import json
from ably import AblyRealtime

ABLY_API_KEY = os.getenv("ABLY_API_KEY")
CHANNEL_NAME = os.getenv("ABLY_CHANNEL", "process-control")

if not ABLY_API_KEY:
    raise ValueError("Set ABLY_API_KEY as an environment variable or in this script.")

async def ably_listener(channel):
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

async def ably_sender(channel):
    loop = asyncio.get_running_loop()
    def sync_input(prompt):
        return loop.run_in_executor(None, input, prompt)
    while True:
        print("\nAbly Async Test Menu:")
        print("1. Send a command event")
        print("2. Send a status query")
        print("3. Exit")
        choice = (await sync_input("Select an option (1-3): ")).strip()
        if choice == '1':
            script = (await sync_input("Script name (e.g. dhanfeed_sync.py): ")).strip()
            action = (await sync_input("Action (start/stop/restart): ")).strip()
            payload = {"action": action, "script": script}
            await channel.publish("command", json.dumps(payload))
            print(f"[SENT] Event: command | Data: {payload}")
        elif choice == '2':
            await channel.publish("command", json.dumps({"action": "status"}))
            print(f"[SENT] Event: command | Data: {{'action': 'status'}}")
        elif choice == '3':
            print("Exiting.")
            break
        else:
            print("Invalid option.")
        await asyncio.sleep(0.5)

async def main():
    ably = AblyRealtime(ABLY_API_KEY)
    channel = ably.channels.get(CHANNEL_NAME)
    # Start listener in background
    listener_task = asyncio.create_task(ably_listener(channel))
    # Run sender in foreground
    await ably_sender(channel)
    # Cancel listener on exit
    listener_task.cancel()
    try:
        await listener_task
    except asyncio.CancelledError:
        pass

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\nExiting Ably async test.")
