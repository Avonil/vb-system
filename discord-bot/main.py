import os
import discord
import aiohttp
import json
import logging
import asyncio
import redis.asyncio as redis
from dotenv import load_dotenv

# Завантажуємо локальний .env
load_dotenv()

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("DiscordBot")

DISCORD_TOKEN = os.getenv("DISCORD_TOKEN")
CORE_API_URL = os.getenv("CORE_API_URL", "http://localhost:3000/api/bot/command")
REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")

class VerisDiscordBot(discord.Client):
    def __init__(self):
        intents = discord.Intents.default()
        intents.message_content = True
        super().__init__(intents=intents)

    async def setup_hook(self):
        # Запускаємо слухач алертів з Redis у фоні
        self.loop.create_task(self.redis_listener())

    async def on_ready(self):
        logger.info(f'✅ Logged in as {self.user}')

    async def on_message(self, message):
        if message.author.bot:
            return

        # Перехоплюємо команди ! або / і кидаємо в Ядро
        if message.content.startswith("!") or message.content.startswith("/"):
            parts = message.content[1:].split()
            if not parts: return
            
            command = parts[0].lower()
            args = parts[1:]
            
            try:
                async with aiohttp.ClientSession() as session:
                    payload = {
                        "platform": "discord",
                        "serverId": str(message.guild.id) if message.guild else "",
                        "user": message.author.name,
                        "command": command,
                        "args": args
                    }
                    async with session.post(CORE_API_URL, json=payload) as resp:
                        if resp.status == 200:
                            data = await resp.json()
                            for text in data.get("responses", []):
                                await message.channel.send(text)
            except Exception as e:
                logger.error(f"❌ HTTP request to Core failed: {e}")

    async def redis_listener(self):
        """Слухає Redis для алертів про початок стріму"""
        try:
            r = redis.from_url(REDIS_URL, decode_responses=True)
            pubsub = r.pubsub()
            await pubsub.subscribe("alerts_discord")
            logger.info("🎧 Subscribed to Redis 'alerts_discord'")
            
            async for msg in pubsub.listen():
                if msg["type"] == "message":
                    data = json.loads(msg["data"])
                    channel_id = int(data.get("channelId", 0))
                    text = data.get("message", "")
                    
                    text = text.replace("{streamer}", data.get("streamer", ""))
                    text = text.replace("{url}", data.get("url", ""))
                    text = text.replace("{category}", data.get("category", ""))
                    text = text.replace("{title}", data.get("title", ""))
                    
                    channel = self.get_channel(channel_id)
                    if channel:
                        await channel.send(text)
        except Exception as e:
            logger.error(f"❌ Redis Listener Error: {e}")

async def main():
    if not DISCORD_TOKEN:
        logger.critical("DISCORD_TOKEN is missing!")
        return
    bot = VerisDiscordBot()
    await bot.start(DISCORD_TOKEN)

if __name__ == "__main__":
    asyncio.run(main())