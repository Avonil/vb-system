import os
import logging
import asyncio
import aiohttp
import json
import redis.asyncio as redis
from telegram import Update
from telegram.ext import ApplicationBuilder, CommandHandler, MessageHandler, filters, ContextTypes
from dotenv import load_dotenv

load_dotenv()

logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(levelname)s - %(message)s')
logger = logging.getLogger("TelegramBot")

TELEGRAM_TOKEN = os.getenv("TELEGRAM_TOKEN")
CORE_API_URL = os.getenv("CORE_API_URL", "http://localhost:3000/api/bot/command")
REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379")

class VerisTelegramBot:
    def __init__(self, token):
        self.app = ApplicationBuilder().token(token).build()
        self.setup_handlers()

    def setup_handlers(self):
        self.app.add_handler(CommandHandler("start", self.cmd_start))
        # Перехоплюємо всі інші команди
        self.app.add_handler(MessageHandler(filters.COMMAND, self.cmd_universal))

    async def cmd_start(self, update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
        await update.message.reply_text(f'Привіт, {update.effective_user.first_name}!')

    async def cmd_universal(self, update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
        if not update.message or not update.message.text: return
            
        text = update.message.text
        if text.startswith("/"):
            parts = text[1:].split()
            command = parts[0].split('@')[0].lower()
            args = parts[1:]
            
            try:
                async with aiohttp.ClientSession() as session:
                    payload = {
                        "platform": "telegram",
                        "serverId": str(update.message.chat_id),
                        "user": update.effective_user.first_name,
                        "command": command,
                        "args": args
                    }
                    async with session.post(CORE_API_URL, json=payload) as resp:
                        if resp.status == 200:
                            data = await resp.json()
                            for response_text in data.get("responses", []):
                                await update.message.reply_text(response_text)
            except Exception as e:
                logger.error(f"❌ HTTP request to Core failed: {e}")

    async def redis_listener(self):
        try:
            r = redis.from_url(REDIS_URL, decode_responses=True)
            pubsub = r.pubsub()
            await pubsub.subscribe("alerts_telegram")
            logger.info("🎧 Subscribed to Redis 'alerts_telegram'")
            
            async for msg in pubsub.listen():
                if msg["type"] == "message":
                    data = json.loads(msg["data"])
                    chat_id = data.get("chatId")
                    text = data.get("message", "")
                    
                    text = text.replace("{streamer}", data.get("streamer", ""))
                    text = text.replace("{url}", data.get("url", ""))
                    text = text.replace("{category}", data.get("category", ""))
                    text = text.replace("{title}", data.get("title", ""))
                    
                    await self.app.bot.send_message(chat_id=chat_id, text=text)
        except Exception as e:
            logger.error(f"❌ Redis Listener Error: {e}")

    async def start(self):
        await self.app.initialize()
        await self.app.start()
        await asyncio.gather(
            self.app.updater.start_polling(),
            self.redis_listener()
        )

async def main():
    if not TELEGRAM_TOKEN:
        logger.critical("TELEGRAM_TOKEN is missing!")
        return
    bot = VerisTelegramBot(TELEGRAM_TOKEN)
    await bot.start()

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logger.info("🛑 Bot stopped.")