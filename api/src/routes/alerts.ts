import { Elysia, t } from 'elysia';
import { jwt } from '@elysiajs/jwt';
import { User } from '../models/User';
import { sendToCore } from '../setup/core';

export const alertsRoutes = new Elysia({ prefix: '/alerts' })
    .use(jwt({ name: 'jwt', secret: process.env.JWT_SECRET! }))
    
    .derive(async ({ jwt, headers }) => {
        const token = headers['authorization']?.replace('Bearer ', '');
        const profile = token ? await jwt.verify(token) : null;
        return { user: profile };
    })
    
    .guard({
        beforeHandle: ({ user, set }) => {
            if (!user) { set.status = 401; return { error: 'Unauthorized' }; }
        }
    })

    // Оновлення налаштувань Telegram та Discord
    .put('/integrations', async ({ user, body, set }) => {
        try {
            const dbUser = await User.findById(user!.id);
            if (!dbUser) { set.status = 404; return { error: 'User not found' }; }

            const { discord, telegram } = body as any;

            if (discord) dbUser.modules.alerts.discord = discord;
            if (telegram) dbUser.modules.alerts.telegram = telegram;
            dbUser.modules.alerts.enabled = true;

            await dbUser.save();

            // 🔥 СИГНАЛ ЯДРУ
            sendToCore('USER_UPDATED', {
                twitchId: dbUser.twitchId,
                username: dbUser.username
            });

            return { success: true, message: 'Integrations updated successfully' };
        } catch (e) {
            set.status = 500;
            return { error: 'Database error' };
        }
    });