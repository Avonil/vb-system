import os
import sys
import logging
from dotenv import load_dotenv

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[logging.StreamHandler(sys.stdout)]
)
logger = logging.getLogger("VerisConfig")

loaded = load_dotenv()
if loaded:
    logger.info("📂 Loaded configuration from .env file")
else:
    logger.warning("⚠️ .env file not found! Relying on system environment variables.")

CORE_WEBSOCKET_URL = os.getenv("CORE_WS_URL", "ws://localhost:9090/ws?type=bot")
BOT_NICK = os.getenv("BOT_USERNAME", "VerisBot")
BOT_TOKEN = os.getenv("TWITCH_AUTH_TOKEN")

# 🔥 ДОДАЛИ НОВІ ЗМІННІ ДЛЯ TWITCHIO V3
CLIENT_ID = os.getenv("TWITCH_CLIENT_ID")
CLIENT_SECRET = os.getenv("TWITCH_CLIENT_SECRET")
BOT_ID = os.getenv("BOT_ID")

if not all([BOT_TOKEN, CLIENT_ID, CLIENT_SECRET, BOT_ID]):
    logger.critical("❌ Error: Missing critical Twitch credentials (TOKEN, CLIENT_ID, SECRET or BOT_ID) in .env!")
    sys.exit(1)

if not BOT_TOKEN.startswith("oauth:"):
    BOT_TOKEN = f"oauth:{BOT_TOKEN}"
    logger.info("🔧 Added 'oauth:' prefix to token automatically.")

logger.info(f"✅ Configuration ready. Bot: {BOT_NICK}")