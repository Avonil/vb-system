from twitchio.ext import commands

class Moderation(commands.Cog):
    def __init__(self, bot):
        self.bot = bot

    @commands.command(name='mute')
    async def cmd_mute(self, ctx: commands.Context, target: str = None, duration: str = "60", *reason_parts):
        if not target: return
        target = target.replace('@', '').lower()
        reason = " ".join(reason_parts) if reason_parts else "Veris Moderation"
        duration_int = int(duration) if duration.isdigit() else 60
        
        try:
            await ctx.send(f"/timeout {target} {duration_int} {reason}")
            await ctx.send(f"🤐 {target} muted for {duration_int}s.")
        except: pass

    @commands.command(name='ban')
    async def cmd_ban(self, ctx: commands.Context, target: str = None, *reason_parts):
        if not target: return
        target = target.replace('@', '').lower()
        reason = " ".join(reason_parts) if reason_parts else "Veris Ban"
        try:
            await ctx.send(f"/ban {target} {reason}")
            await ctx.send(f"🔨 {target} banned.")
        except: pass

def prepare(bot):
    bot.add_cog(Moderation(bot))