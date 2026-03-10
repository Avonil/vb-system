import asyncio
import json
import websockets
import logging
from configs import CORE_WEBSOCKET_URL

logger = logging.getLogger("WS_Client")

class WebSocketClient:
    def __init__(self, bot_service=None):
        self.websocket = None
        self.bot_service = bot_service

    def set_bot(self, bot):
        self.bot_service = bot

    async def connect(self):
        while True:
            try:
                logger.info(f"🔌 Connecting to Core at {CORE_WEBSOCKET_URL}...")
                async with websockets.connect(CORE_WEBSOCKET_URL) as ws:
                    self.websocket = ws
                    
                    # 🔥 1. ВІДПРАВЛЯЄМО HANDSHAKE! Ядро має знати, що ми Твіч-бот
                    handshake = {
                        "type": "HANDSHAKE",
                        "client": "twitch-bot",
                        "data": {"role": "bot", "platform": "twitch"}
                    }
                    await ws.send(json.dumps(handshake))
                    logger.info("✅ Connected to Core WebSocket and sent Handshake.")
                    
                    await self.listen_for_messages()
            except Exception as e:
                logger.error(f"❌ WS Connection error: {e}")
                await asyncio.sleep(5)

    async def listen_for_messages(self):
        async for message in self.websocket:
            try:
                logger.info(f"📥 RAW WS MSG RECV: {message}")
                ws_message = json.loads(message)
                await self.handle_message(ws_message)
            except Exception as e:
                logger.error(f"❌ Error processing message: {e}")

    async def send_command_to_core(self, channel, user, command, args):
        if not self.websocket: return
        payload = {
            "type": "CHAT_COMMAND",
            "data": {
                "platform": "twitch",
                "channel": channel,
                "user": user,
                "command": command,
                "args": args
            }
        }
        await self.websocket.send(json.dumps(payload))
        logger.info(f"📤 Sent command '!{command}' to Core.")

    async def handle_message(self, ws_message):
        msg_type = ws_message.get("type")
        # Іноді Ядро присилає дані прямо в тілі, іноді в "data"
        msg_data = ws_message.get("data", ws_message) 

        # 🔥 2. СЛУХАЄМО І JOIN_CHANNEL, І USER_UPDATED!
        if msg_type in ["JOIN_CHANNEL", "USER_UPDATED"] and self.bot_service:
            # Шукаємо нікнейм де тільки можна
            target_channel = msg_data.get("username") or msg_data.get("twitch_username") or msg_data.get("channel")
            if target_channel:
                logger.info(f"🎯 Catch event '{msg_type}'. Joining channel: {target_channel}")
                asyncio.create_task(self.bot_service.safe_join(target_channel))
            else:
                logger.warning(f"⚠️ Received {msg_type} but no username found in payload: {msg_data}")

        elif msg_type == "SEND_MESSAGE" and self.bot_service:
            channel = msg_data.get("channel")
            text = msg_data.get("text")
            if channel and text:
                asyncio.create_task(self.bot_service.send_to_channel(channel, text))

    async def send_log(self, channel, user, content):
        if not self.websocket: return
        payload = {
            "type": "CHAT_LOG",
            "data": {
                "platform": "twitch",
                "channel": channel,
                "user": user,
                "message": content
            }
        }
        await self.websocket.send(json.dumps(payload))