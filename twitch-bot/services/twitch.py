import asyncio
import logging
from twitchio.ext import commands
from configs import BOT_TOKEN, BOT_NICK

logger = logging.getLogger("TwitchCore")

class VerisServiceBot(commands.Bot):
    def __init__(self, ws_client):
        # Робимо нормально: бот на старті заходить тільки у свій власний чат (з .env)
        # Усі інші канали йому буде динамічно підкидати Go Ядро через вебсокет!
        start_channel = BOT_NICK.lower()
        
        super().__init__(
            token=BOT_TOKEN,
            prefix='!',
            initial_channels=[start_channel] 
        )
        self.ws_client = ws_client
        self.joined_channels_set = set([start_channel])
        self.is_twitch_ready = False
        logger.info(f"🤖 Veris Service Bot initialized as {BOT_NICK}")

    async def event_ready(self):
        self.is_twitch_ready = True
        logger.info(f"✅ Logged in as | {self.nick} | User ID: {self.user_id}")

    async def event_message(self, message):
        if message.echo:
            logger.info(f"[{message.channel.name}] {message.author.name}: {message.content}")
            return

        logger.info(f"[{message.channel.name}] {message.author.name}: {message.content}")

        await self.ws_client.send_log(
            channel=message.channel.name,
            user=message.author.name,
            content=message.content
        )

        if message.content.startswith('!'):
            parts = message.content.split()
            if not parts: return
            command = parts[0][1:].lower()
            args = parts[1:]
            
            await self.ws_client.send_command_to_core(
                channel=message.channel.name,
                user=message.author.name,
                command=command,
                args=args
            )

        await self.handle_commands(message)

    async def send_to_channel(self, channel_name, text):
        if not self.is_twitch_ready:
            return
        channel = self.get_channel(channel_name)
        if channel:
            await channel.send(text)

    async def event_command_error(self, ctx, error):
        if isinstance(error, commands.CommandNotFound):
            # Ігноруємо помилку "команду не знайдено", 
            # бо ми все одно відправили її в Ядро для перевірки в БД!
            pass
        else:
            # Інші реальні помилки логуємо
            logger.error(f"❌ Command Error in {ctx.command}: {error}")
    
    async def safe_join(self, channel_name):
        """Динамічне підключення до нових стрімерів (викликається Ядром)"""
        channel_name = channel_name.lower()
        while not self.is_twitch_ready:
            await asyncio.sleep(1)
            
        if channel_name in self.joined_channels_set:
            return

        try:
            await asyncio.sleep(0.5) 
            await self.join_channels([channel_name])
            self.joined_channels_set.add(channel_name)
            logger.info(f"✅ SUCCESS: Dynamically joined streamer channel -> {channel_name}")
        except Exception as e:
            logger.error(f"❌ Failed to join {channel_name}: {e}")