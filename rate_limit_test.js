import http from 'k6/http';
import { check } from 'k6';

export const options = {
    scenarios: {
        statis_rate_limit: {
            executor: 'shared-iterations',
            vus: 1,             
            iterations: 20,   
            maxDuration: '10s', 
        },
    },
};

export default function () {
    const url = 'http://localhost:8080/api/v1/auth/login';
    
    const payload = JSON.stringify({
        email: 'ekokurniawaann@gmail.com',
        password: 'password_salah'
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'Status 400 (Kredensial Salah - Lolos Limiter)': (r) => r.status === 400,
        'Status 429 (Terblokir - Distributed Redis Limiter Working)': (r) => r.status === 429,
    });
}