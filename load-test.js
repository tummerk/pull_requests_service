import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Trend } from 'k6/metrics';

// --- МЕТРИКИ ДЛЯ ОТЧЕТА ---
const teamGetDuration = new Trend('team_get_duration');
const userReviewsDuration = new Trend('user_reviews_duration');
const statsGetDuration = new Trend('stats_get_duration');

// --- НАСТРОЙКИ ТЕСТА ---
export const options = {
  scenarios: {
    read_only: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '10s', target: 0 },
      ],
      gracefulRampDown: '15s',
    },
  },
  thresholds: {
    'http_req_failed': ['rate<0.001'],
    'http_req_duration': ['p(95)<300'],
  },
};

// --- ДАННЫЕ ДЛЯ ТЕСТА (С ТВОИМИ РЕАЛЬНЫМИ ID) ---

// Я выбрал несколько команд из твоего списка
const TEAMS_TO_TEST = [
  'loadtest_team_4_0',
  'loadtest_team_11_1',
  'loadtest_team_2_0',
  'loadtest_team_10_0',
  'loadtest_team_9_2',
  'loadtest_team_15_0',
  'loadtest_team_7_4',
  'loadtest_team_1_0',
];

// И несколько пользователей
const USERS_TO_TEST = [
  'u_4_0_1',
  'u_3_0_2',
  'u_11_1_1',
  'u_8_1_2',
  'u_2_0_1',
  'u_10_1_2',
  'u_9_0_1',
  'u_6_1_2',
];


// --- ОСНОВНАЯ ЛОГИКА ТЕСТА ---
export default function () {
  const BASE_URL = 'http://localhost:8080';

  group('Get Team Info', function () {
    const teamName = TEAMS_TO_TEST[Math.floor(Math.random() * TEAMS_TO_TEST.length)];
    const res = http.get(`${BASE_URL}/team/get?team_name=${teamName}`);
    teamGetDuration.add(res.timings.duration);
    check(res, {
      'Get Team: status is 200': (r) => r.status === 200,
    });
  });

  sleep(0.5);

  group('Get User Reviews', function () {
    const userId = USERS_TO_TEST[Math.floor(Math.random() * USERS_TO_TEST.length)];
    const res = http.get(`${BASE_URL}/users/getReview?user_id=${userId}`);
    userReviewsDuration.add(res.timings.duration);
    check(res, {
      'Get User Reviews: status is 200': (r) => r.status === 200,
    });
  });

  sleep(0.5);

  group('Get Global Stats', function () {
    const res = http.get(`${BASE_URL}/user_stats`);
    statsGetDuration.add(res.timings.duration);
    check(res, {
      'Get Stats: status is 200': (r) => r.status === 200,
    });
  });
  
  sleep(1);
}