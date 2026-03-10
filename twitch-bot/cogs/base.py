from twitchio.ext import commands

class General(commands.Cog):
    def __init__(self, bot):
        self.bot = bot

    @commands.command(name='ping')
    async def cmd_ping(self, ctx: commands.Context):
        await ctx.send(f"Pong! 🏓 Veris ({self.bot.nick}) is active in {ctx.channel.name}!")

def prepare(bot):
    bot.add_cog(General(bot))