import dotenv from 'dotenv';
import { Elysia } from 'elysia';
import { cors } from '@elysiajs/cors';
import { swagger } from '@elysiajs/swagger';
import { connectDB } from './setup/db';
import { connectCore } from './setup/core';
import { authRoutes } from './routes/auth';
import { commandRoutes } from './routes/commands';
import { alertsRoutes } from './routes/alerts';

dotenv.config();

// 1. Запускаємо підключення
// await connectDB();
// connectCore();

// 2. Ініціалізуємо API
const app = new Elysia()
  // Додаємо плагіни
  .use(cors()) // Дозволяє фронту стукатись
  .use(swagger({
    path: '/docs',
    documentation: {
        info: {
            title: 'Veris API',
            version: '1.0.0'
        }
    }
  }))

  .use(authRoutes)
  .use(commandRoutes)
  .use(alertsRoutes)
  
  // 3. Базовий роут (Health Check)
  .get('/', () => ({ 
    status: 'online', 
    service: 'Veris API', 
    timestamp: new Date().toISOString() 
  }))

  // 4. Приклад роута: Ping Core
  // Тільки для тесту! Потім видалимо.
  .get('/test-core', () => {
      // Тут ми використовуємо нашу функцію відправки
      const { sendToCore } = require('./setup/core');
      sendToCore('TEST_PING', { message: 'Hello from API Endpoint!' });
      return { success: true, message: 'Sent ping to Core' };
  })

  // 5. Запуск сервера
  .listen({
    port: process.env.API_PORT || 3000,
    hostname: '0.0.0.0',
    tls: {
        cert: Bun.file("server.crt"), // Шлях до твого сертифікату
        key: Bun.file("server.key"),  // Шлях до ключа
    }
  });

console.log(
  `🦊 Veris API is running at http://${app.server?.hostname}:${app.server?.port}`
);