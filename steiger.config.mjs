import { defineConfig } from 'steiger';
import fsd from '@feature-sliced/steiger-plugin';

// Проверка правил Feature-Sliced Design: npm run lint:fsd
export default defineConfig([
  ...fsd.configs.recommended,
  {
    // Сущности повторяют модули бэкенда (каталог, маршруты, магазин…) и понадобятся
    // фичам и слою API, даже если сейчас их использует одна страница.
    files: ['./src/entities/**'],
    rules: { 'fsd/insignificant-slice': 'off' },
  },
]);
