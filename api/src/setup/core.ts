// Глобальна змінна для сокета, щоб ми могли слати меседжі з контролерів
export let coreSocket: WebSocket | null = null;

const rconInterval = Number(process.env.RECONNECT_INTERVAL);

export const connectCore = () => {
  const url = process.env.CORE_WS_URL;
  if (!url) {
    console.error('❌ CORE_WS_URL is missing!');
    return;
  }

  console.log(`🔌 Connecting to Core at ${url}...`);
  
  // Bun має вбудований WebSocket клієнт!
  coreSocket = new WebSocket(url);

  coreSocket.onopen = () => {
    console.log('✅ Connected to Veris Core!');
    // Можна відразу відправити "Привіт" ядру
    coreSocket?.send(JSON.stringify({ type: 'API_HANDSHAKE', data: { status: 'ok' } }));
  };

  coreSocket.onmessage = (event) => {
    console.log('📩 Message from Core:', event.data);
  };

  coreSocket.onclose = () => {
    console.warn('⚠️ Disconnected from Core. Reconnecting...');
    coreSocket = null;
    setTimeout(connectCore, isNaN(rconInterval) ? 3000 : rconInterval);
  };

  coreSocket.onerror = (error) => {
    console.error('❌ Core WebSocket Error:', error);
    coreSocket?.close();
  };
};

// Функція-хелпер, щоб відправляти команди на ядро з будь-якого місця коду
export const sendToCore = (type: string, data: any) => {
    if (coreSocket && coreSocket.readyState === WebSocket.OPEN) {
        coreSocket.send(JSON.stringify({ type, data }));
    } else {
        console.warn('⚠️ Cannot send to Core: Disconnected');
    }
}