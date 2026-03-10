# Для Твіч-бота
docker build --platform linux/amd64 -t avonill/veris-twitch-bot:latest ./twitch-bot
docker push avonill/veris-twitch-bot:latest

# Для Телеграм-бота
docker build --platform linux/amd64 -t avonill/veris-telegram-bot:latest ./telegram-bot
docker push avonill/veris-telegram-bot:latest

# Для Дискорд-бота
docker build --platform linux/amd64 -t avonill/veris-discord-bot:latest ./discord-bot
docker push avonill/veris-discord-bot:latest

# Для Ядра (якщо воно теж почне падати)
docker build --platform linux/amd64 -t avonill/veris-core:latest ./core
docker push avonill/veris-core:latest

docker build --platform linux/amd64 -t avonill/veris-api:latest ./api
docker push avonill/veris-api:latest