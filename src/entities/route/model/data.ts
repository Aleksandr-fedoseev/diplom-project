import type { RouteItem } from './types';

export const ROUTES: RouteItem[] = [
  {
    id: 'nikolskaya-street',
    slug: 'nikolskaya-ulitsa',
    title: 'Главная купеческая улица: Никольская и Биржа',
    description: 'Пешеходный маршрут по парадной улице Косы: Здание Биржи, банк Лидваля, дома Агамжановой и Тавризова.',
    coverImage: '/heritage/birzha/cover.jpg',
    duration: '45 минут',
    distance: '1.2 км',
    stopsCount: 4,
    objectIds: ['birzha', 'azovo-donskoy-bank', 'dom-agamzhanovoy', 'dom-tavrizova'],
  },
  {
    id: 'kutum-manors',
    slug: 'usadby-reki-kutum',
    title: 'Купеческие особняки набережной Кутума',
    description: 'Осмотр усадеб Губина, Шелехова, Будагова и первой городской электростанции.',
    coverImage: '/heritage/dom-gubina/cover.jpg',
    duration: '1 час',
    distance: '2.0 км',
    stopsCount: 4,
    objectIds: ['dom-gubina', 'elektrostantsiya', 'dom-budagova', 'dom-shelehova'],
  },
];
