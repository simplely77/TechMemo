import axios from 'axios';
import { refreshToken } from '@/services/authService';

const instance = axios.create({
    baseURL: 'http://localhost:8080/api/v1',
    withCredentials: true,
    headers: { 'Content-Type': 'application/json', }
});

instance.interceptors.request.use(config => {
    const accessToken = localStorage.getItem('access_token');
    if (accessToken) config.headers['Authorization'] = `Bearer ${accessToken}`;
    return config;
});

instance.interceptors.response.use(
    res => res,
    async err => {
        const originalRequest = err.config

        if (err.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;
            try {
                const rt = localStorage.getItem('refresh_token');
                if (!rt) {
                    // 没有refreshToken，清除本地存储并跳转到登录
                    localStorage.removeItem('access_token');
                    localStorage.removeItem('refresh_token');
                    window.location.href = '/login';
                    return Promise.reject(err);
                }

                const res = await refreshToken({
                    refresh_token: rt,
                });

                if (res) {
                    const { access_token, refresh_token } = res;
                    localStorage.setItem('access_token', access_token);
                    localStorage.setItem('refresh_token', refresh_token);

                    // 更新原请求的token
                    originalRequest.headers['Authorization'] = `Bearer ${access_token}`;
                    return instance(originalRequest);
                }
            } catch (refreshErr) {
                // 刷新失败，清除本地存储并跳转到登录
                localStorage.removeItem('access_token');
                localStorage.removeItem('refresh_token');
                window.location.href = '/login';
                return Promise.reject(refreshErr);
            }
        }
        return Promise.reject(err);
    }
)

export default instance;