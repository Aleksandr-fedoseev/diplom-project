'use client';

import { useCallback } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { setToken, useToken } from '@/shared/api';
import { sessionApi, sessionKeys } from '../api/sessionApi';
import { roleAtLeast, type Role, type User } from '@/entities/user/@x/session';

// Текущий пользователь. Роль и блокировка берутся с сервера (/auth/me), а не из токена:
// бэкенд и сам читает их из базы при каждом запросе.
export function useSession() {
  const token = useToken();
  const queryClient = useQueryClient();
  const me = useQuery({
    queryKey: [...sessionKeys.me, token],
    queryFn: sessionApi.me,
    enabled: !!token,
    staleTime: 60_000,
    retry: false,
  });

  const login = useCallback(
    async (email: string, password: string) => {
      const { token, user } = await sessionApi.login(email, password);
      queryClient.setQueryData([...sessionKeys.me, token], user);
      setToken(token);
      return user;
    },
    [queryClient],
  );

  const logout = useCallback(() => {
    setToken(null);
    queryClient.clear(); // корзина, заказы и заявки прежнего пользователя не должны остаться в кэше
  }, [queryClient]);

  // Профиль изменён — обновляем текущего пользователя без нового запроса.
  const setUser = useCallback(
    (u: User) => queryClient.setQueryData([...sessionKeys.me, token], u),
    [queryClient, token],
  );

  const user = token ? me.data : undefined;
  return {
    user,
    // Пока токен есть, а ответа /auth/me ещё нет, статус неизвестен — не показываем «войдите».
    isLoading: !!token && me.isPending,
    isAuthed: !!user,
    hasRole: (min: Role) => roleAtLeast(user?.role, min),
    login,
    logout,
    setUser,
  };
}
