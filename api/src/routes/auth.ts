import { Elysia, t } from "elysia";
import { jwt } from "@elysiajs/jwt";
import { getTwitchToken, getTwitchUser } from "../utils/twitch";
import { User } from "../models/User";
import { sendToCore } from "../setup/core";

// 🔥 Ті самі права, що були в Go (ALL PERMISSIONS)
const TWITCH_SCOPES = [
  // ... (весь список скоупів, що був раніше - залишаємо без змін) ...
  "analytics:read:extensions",
  "analytics:read:games",
  "bits:read",
  "channel:edit:commercial",
  "channel:manage:broadcast",
  "channel:manage:extensions",
  "channel:manage:moderators",
  "channel:manage:polls",
  "channel:manage:predictions",
  "channel:manage:raids",
  "channel:manage:redemptions",
  "channel:manage:schedule",
  "channel:manage:videos",
  "channel:manage:vips",
  "channel:read:ads",
  "channel:read:charity",
  "channel:read:editors",
  "channel:read:goals",
  "channel:read:guest_star",
  "channel:read:hype_train",
  "channel:read:polls",
  "channel:read:predictions",
  "channel:read:redemptions",
  "channel:read:stream_key",
  "channel:read:subscriptions",
  "channel:read:vips",
  "channel:moderate",
  "channel:bot",
  "chat:edit",
  "chat:read",
  "user:write:chat",
  "user:bot",
  "clips:edit",
  "moderator:manage:announcements",
  "moderator:manage:automod",
  "moderator:manage:automod_settings",
  "moderator:manage:banned_users",
  "moderator:manage:blocked_terms",
  "moderator:manage:chat_messages",
  "moderator:manage:chat_settings",
  "moderator:manage:guest_star",
  "moderator:manage:shield_mode",
  "moderator:manage:shoutouts",
  "moderator:manage:unban_requests",
  "moderator:manage:warnings",
  "moderator:read:automod_settings",
  "moderator:read:banned_users",
  "moderator:read:blocked_terms",
  "moderator:read:chat_settings",
  "moderator:read:chatters",
  "moderator:read:followers",
  "moderator:read:guest_star",
  "moderator:read:shield_mode",
  "moderator:read:shoutouts",
  "moderator:read:suspicious_users",
  "moderator:read:unban_requests",
  "moderator:read:warnings",
  "user:edit",
  "user:edit:broadcast",
  "user:edit:follows",
  "user:manage:blocked_users",
  "user:manage:chat_color",
  "user:manage:whispers",
  "user:read:blocked_users",
  "user:read:broadcast",
  "user:read:email",
  "user:read:follows",
  "user:read:moderated_channels",
  "user:read:subscriptions",
  "user:read:emotes",
  "whispers:edit",
  "whispers:read",
].join(" ");

export const authRoutes = new Elysia({ prefix: "/auth" })
  .use(jwt({ name: "jwt", secret: process.env.JWT_SECRET! }))

  // 1. Логін СТРІМЕРА
  .get("/login", ({ redirect }) => {
    const url = `https://id.twitch.tv/oauth2/authorize?client_id=${process.env.TWITCH_CLIENT_ID}&redirect_uri=${process.env.TWITCH_REDIRECT_URI}&response_type=code&scope=${TWITCH_SCOPES}&state=streamer`;
    return redirect(url);
  })

  // 2. Логін БОТА (state=bot)
  .get("/bot-login", ({ redirect }) => {
    const url = `https://id.twitch.tv/oauth2/authorize?client_id=${process.env.TWITCH_CLIENT_ID}&redirect_uri=${process.env.TWITCH_REDIRECT_URI}&response_type=code&scope=${TWITCH_SCOPES}&state=bot`;
    return redirect(url);
  })

  // 3. Callback
  .get("/callback", async ({ query, jwt, set }) => {
    const code = query.code as string;
    const state = query.state as string;

    if (!code) return { error: "No code provided" };

    try {
      // А. Обмін коду на токен
      const tokens = await getTwitchToken(code);
      const twitchUser = await getTwitchUser(tokens.access_token);

      console.log(tokens);
      console.log(twitchUser);

      console.log({
        tokens: tokens,
        user: twitchUser,
      });

      // === ГІЛКА БОТА ===
      if (state !== "bot") {
        return {
          status: "Authenticated Successfully!",
        };
      }

      // === ГІЛКА СТРІМЕРА ===
      console.log(`👤 STREAMER Login: ${twitchUser.login}`);

      // 1. Шукаємо або створюємо юзера
      let user = await User.findOne({ twitchId: twitchUser.id });

      if (!user) {
        console.log("🆕 New User! Creating Passport...");
        user = new User({
          twitchId: twitchUser.id,
          username: twitchUser.login,
          displayName: twitchUser.display_name,
          avatarUrl: twitchUser.profile_image_url,
          role: "streamer",
          isStreamer: true,

          // Заповнюємо Auth
          auth: {
            accessToken: tokens.access_token,
            refreshToken: tokens.refresh_token,
            scopes: tokens.scope,
            expiresAt: new Date(Date.now() + tokens.expires_in * 1000),
          },

          // Дефолтний конфіг заповниться сам через Mongoose default values,
          // але модулі краще ініціалізувати явно
          modules: {
            chat: {
              enabled: true,
              announceLive: true,
              customCommands: [
                {
                  trigger: "!veris",
                  response: "Veris Bot is active! 🚀",
                  cooldown: 10,
                  userLevel: "everyone",
                  enabled: true,
                },
              ],
            },
            moderation: { enabled: true, bannedWords: [] },
            alerts: { enabled: false },
          },
        });
      } else {
        console.log("🔄 Existing User. Updating tokens...");
        // Оновлюємо інфо
        user.username = twitchUser.login;
        user.displayName = twitchUser.display_name;
        user.avatarUrl = twitchUser.profile_image_url;
        user.lastLogin = new Date();

        // Оновлюємо токени
        user.auth.accessToken = tokens.access_token;
        user.auth.refreshToken = tokens.refresh_token;
        user.auth.expiresAt = new Date(Date.now() + tokens.expires_in * 1000);
      }

      await user.save();

      // 2. 🔥 СИГНАЛ ЯДРУ (Real-Time Update)
      // Ми кажемо: "Онови цього юзера в пам'яті!"
      sendToCore("USER_UPDATED", {
        twitchId: user.twitchId,
        username: user.username,
      });

      // 3. Генеруємо JWT для фронта
      const sessionToken = await jwt.sign({
        id: user._id.toString(),
        twitchId: user.twitchId,
        role: user.role,
      });

      return {
        success: true,
        user: user.username,
        token: sessionToken,
        message: "Passport Updated & Core Notified!",
      };
    } catch (error) {
      console.error(error);
      set.status = 500;
      return { error: "Authentication failed", details: error };
    }
  });
