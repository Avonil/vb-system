// src/services/api.js
import axios from 'axios';

const API_URL = 'http://185.183.92.97:3000'; // Твій сервер на Netcup

export const connectTwitch = () => {
  // Просто відкриваємо вікно авторизації
  window.location.href = `${API_URL}/auth/twitch`;
};

export const updateCommand = async (commandData) => {
  return await axios.post(`${API_URL}/api/user/commands`, commandData);
};

export const getDiscordInvite = async () => {
  const res = await axios.get(`${API_URL}/api/auth/discord-link`);
  return res.data.url;
};