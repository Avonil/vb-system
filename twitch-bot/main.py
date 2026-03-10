import asyncio
import logging
from services.ws import WebSocketClient
from services.twitch import VerisServiceBot

COGS = [
    'cogs.base',
    'cogs.moder'
]

async def main():
    ws_client = WebSocketClient()
    bot = VerisServiceBot(ws_client)
    ws_client.set_bot(bot)

    # У TwitchIO 2.x це працює без await
    for cog in COGS:
        bot.load_module(cog)
        logging.info(f"⚙️ Loaded module: {cog}")

    loop = asyncio.get_event_loop()
    loop.create_task(ws_client.connect())
    
    logging.info("🔥 Starting Veris Bot Service...")
    await bot.start()

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.info("🛑 Bot stopped by user.")