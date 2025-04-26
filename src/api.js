// src/api.js
import axios from 'axios';
import { API_BASE_URL, API_TOKEN } from './config';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    Authorization: API_TOKEN ? `Bearer ${API_TOKEN}` : undefined,
    'Content-Type': 'application/json',
  },
});

export default api;
