import { Elysia, t } from 'elysia';
import { jwt } from '@elysiajs/jwt';
import { User } from '../models/User';
import { sendToCore } from '../setup/core';

// Схема валідації вхідних даних (щоб не пхали дурниці)
const CommandSchema = t.Object({
    trigger: t.String(),
    response: t.String(),
    cooldown: t.Optional(t.Number({ default: 10 })),
    userLevel: t.Optional(t.String({ default: 'everyone' })), // everyone, sub, mod
    enabled: t.Optional(t.Boolean({ default: true }))
});

export const commandRoutes = new Elysia({ prefix: '/commands' })
    // Підключаємо JWT, бо це захищений роут
    .use(jwt({ name: 'jwt', secret: process.env.JWT_SECRET! }))
    
    // Мідлвар для перевірки токена (User Guard)
    .derive(async ({ jwt, headers, set }) => {
        const authHeader = headers['authorization'];
        if (!authHeader) return { user: null };

        const token = authHeader.startsWith('Bearer ') ? authHeader.slice(7) : authHeader;
        const profile = await jwt.verify(token);

        if (!profile) return { user: null };
        return { user: profile };
    })
    
    // Захист: Якщо юзера немає — відфутболюємо
    .guard({
        beforeHandle: ({ user, set }) => {
            if (!user) {
                set.status = 401;
                return { error: 'Unauthorized' };
            }
        }
    })

    // 1. ОТРИМАТИ ВСІ КОМАНДИ (GET /commands)
    .get('/', async ({ user, set }) => {
        try {
            // Шукаємо юзера в базі
            const dbUser = await User.findById(user!.id);
            if (!dbUser) { set.status = 404; return { error: 'User not found' }; }

            return { 
                success: true, 
                commands: dbUser.modules.chat.customCommands 
            };
        } catch (e) {
            set.status = 500;
            return { error: 'Database error' };
        }
    })

    // 2. ДОДАТИ / ОНОВИТИ КОМАНДУ (PUT /commands)
    .post('/', async ({ user, body, set }) => {
        try {
            const dbUser = await User.findById(user!.id);
            if (!dbUser) { set.status = 404; return { error: 'User not found' }; }

            const newCmd = body;
            // Нормалізуємо тригер (додаємо !, якщо забули, і переводимо в нижній регістр)
            if (!newCmd.trigger.startsWith('!')) newCmd.trigger = '!' + newCmd.trigger;
            newCmd.trigger = newCmd.trigger.toLowerCase();

            // Шукаємо, чи є вже така команда
            const existingIndex = dbUser.modules.chat.customCommands.findIndex(
                c => c.trigger === newCmd.trigger
            );

            if (existingIndex >= 0) {
                // ОНОВЛЮЄМО
                dbUser.modules.chat.customCommands[existingIndex] = {
                    ...dbUser.modules.chat.customCommands[existingIndex],
                    ...newCmd
                };
            } else {
                // СТВОРЮЄМО НОВУ
                dbUser.modules.chat.customCommands.push(newCmd);
            }

            // Зберігаємо в Монгу
            await dbUser.save();

            // 🔥 СИГНАЛ ЯДРУ: "Онови кеш!"
            sendToCore('COMMANDS_UPDATED', {
                twitchId: dbUser.twitchId,
                commands: dbUser.modules.chat.customCommands
            });

            return { success: true, message: 'Command saved', command: newCmd };

        } catch (e) {
            console.error(e);
            set.status = 500;
            return { error: 'Failed to save command' };
        }
    }, {
        body: CommandSchema // Валідація
    })

    // 3. ВИДАЛИТИ КОМАНДУ (DELETE /commands/:trigger)
    .delete('/:trigger', async ({ user, params, set }) => {
        try {
            const dbUser = await User.findById(user!.id);
            if (!dbUser) { set.status = 404; return { error: 'User not found' }; }

            const triggerToDelete = decodeURIComponent(params.trigger).toLowerCase();

            // Фільтруємо масив
            const initialLength = dbUser.modules.chat.customCommands.length;
            dbUser.modules.chat.customCommands = dbUser.modules.chat.customCommands.filter(
                c => c.trigger !== triggerToDelete
            );

            if (dbUser.modules.chat.customCommands.length === initialLength) {
                set.status = 404;
                return { error: 'Command not found' };
            }

            await dbUser.save();

            // 🔥 СИГНАЛ ЯДРУ
            sendToCore('COMMANDS_UPDATED', {
                twitchId: dbUser.twitchId,
                commands: dbUser.modules.chat.customCommands
            });

            return { success: true, message: 'Command deleted' };

        } catch (e) {
            set.status = 500;
            return { error: 'Failed to delete command' };
        }
    });