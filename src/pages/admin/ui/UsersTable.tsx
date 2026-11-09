'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { canBlock, canChangeRole, ROLE_LABELS, UserAvatar, userApi, userKeys, type Role, type User } from '@/entities/user';
import { cn, formatDateTime } from '@/shared/lib';
import { ErrorMessage, QueryState, SearchInput } from '@/shared/ui';

const ROLES: Role[] = ['user', 'moderator', 'admin'];

// Управление пользователями и ролями — прецедент администратора: назначить или снять роль
// модератора, заблокировать. Себе роль не меняет и себя не блокирует (правила entities/user).
export function UsersTable({ actor }: { actor: User }) {
  const qc = useQueryClient();
  const [query, setQuery] = useState('');
  const users = useQuery({ queryKey: userKeys.list, queryFn: userApi.list });
  const onSuccess = () => qc.invalidateQueries({ queryKey: userKeys.list });
  const setRole = useMutation({
    mutationFn: (v: { id: number; role: Role }) => userApi.setRole(v.id, v.role),
    onSuccess,
  });
  const setBlocked = useMutation({
    mutationFn: (v: { id: number; blocked: boolean }) => userApi.setBlocked(v.id, v.blocked),
    onSuccess,
  });
  const error = setRole.error ?? setBlocked.error;

  return (
    <div className="space-y-4">
      <SearchInput placeholder="Поиск по имени или почте" value={query} onChange={setQuery} />
      {error && <ErrorMessage error={error} />}
      <QueryState query={users}>
        {(list) => {
          const q = query.trim().toLowerCase();
          const shown = list.filter((u) => !q || u.name.toLowerCase().includes(q) || u.email.toLowerCase().includes(q));
          return (
            <div className="overflow-x-auto bg-white border border-slate-200 rounded-xl">
              <table className="w-full text-sm">
                <thead className="bg-slate-50 text-xs text-slate-500 text-left">
                  <tr>
                    <th className="p-3">Пользователь</th>
                    <th className="p-3">Роль</th>
                    <th className="p-3">Состояние</th>
                    <th className="p-3" />
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {shown.map((u) => (
                    <tr key={u.id} className={cn(u.isBlocked && 'bg-rose-50/50')}>
                      <td className="p-3 flex items-center gap-3">
                        <UserAvatar user={u} className="w-8 h-8 text-xs" />
                        <div>
                          <div className="font-semibold text-slate-900">
                            {u.name}
                            {u.id === actor.id && <span className="text-xs text-slate-400"> (вы)</span>}
                          </div>
                          <div className="text-xs text-slate-500">
                            {u.email} · с {formatDateTime(u.createdAt)}
                          </div>
                        </div>
                      </td>
                      <td className="p-3">
                        {canChangeRole(actor, u) ? (
                          <select
                            aria-label={`Роль ${u.email}`}
                            value={u.role}
                            disabled={setRole.isPending}
                            onChange={(e) =>
                              setRole.mutate({
                                id: u.id,
                                role: e.target.value as Role,
                              })
                            }
                            className="bg-white border border-slate-200 rounded-lg px-2 py-1 text-xs"
                          >
                            {ROLES.map((r) => (
                              <option key={r} value={r}>
                                {ROLE_LABELS[r]}
                              </option>
                            ))}
                          </select>
                        ) : (
                          <span className="text-xs">{ROLE_LABELS[u.role]}</span>
                        )}
                      </td>
                      <td className="p-3 text-xs">
                        {u.isBlocked ? <span className="font-bold text-rose-700">Заблокирован</span> : 'Активен'}
                      </td>
                      <td className="p-3 text-right">
                        {canBlock(actor, u) && (
                          <button
                            type="button"
                            disabled={setBlocked.isPending}
                            onClick={() =>
                              confirm(`${u.isBlocked ? 'Разблокировать' : 'Заблокировать'} ${u.email}?`) &&
                              setBlocked.mutate({
                                id: u.id,
                                blocked: !u.isBlocked,
                              })
                            }
                            className={cn(
                              'px-3 py-1 rounded-lg text-xs font-bold',
                              u.isBlocked ? 'bg-slate-100 hover:bg-slate-200' : 'bg-rose-100 text-rose-800 hover:bg-rose-200',
                            )}
                          >
                            {u.isBlocked ? 'Разблокировать' : 'Заблокировать'}
                          </button>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          );
        }}
      </QueryState>
    </div>
  );
}
