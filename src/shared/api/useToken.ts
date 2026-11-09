'use client';

import { useSyncExternalStore } from 'react';
import { getToken, subscribeToken } from './token';

// Токен как состояние React: компонент перерисуется при входе и выходе.
export const useToken = () => useSyncExternalStore(subscribeToken, getToken, () => null);
