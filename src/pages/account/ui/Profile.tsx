'use client';

import { useRef, useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Camera, CheckCircle2 } from 'lucide-react';
import { sessionApi, useSession } from '@/entities/session';
import { ROLE_LABELS, UserAvatar, type User } from '@/entities/user';
import { formatDateTime } from '@/shared/lib';
import { Button, ErrorMessage, Field, Input } from '@/shared/ui';

const AVATAR_MAX_MB = 2; // как на сервере (auth.MaxAvatarBytes)

const card = 'bg-white border border-slate-200 rounded-2xl p-6 space-y-4';
const smallButton = 'px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wider disabled:opacity-50';

function Saved({ show }: { show: boolean }) {
  if (!show) return null;
  return (
    <span className="inline-flex items-center gap-1 text-xs font-semibold text-emerald-700">
      <CheckCircle2 className="w-4 h-4" /> Сохранено
    </span>
  );
}

function AvatarCard({ user }: { user: User }) {
  const { setUser } = useSession();
  const input = useRef<HTMLInputElement>(null);
  const [error, setError] = useState<unknown>(null);
  const upload = useMutation({ mutationFn: sessionApi.setAvatar, onSuccess: setUser, onError: setError });
  const remove = useMutation({ mutationFn: sessionApi.removeAvatar, onSuccess: setUser, onError: setError });

  return (
    <section className={card}>
      <h2 className="text-base font-bold text-slate-900">Аватар</h2>
      <div className="flex items-center gap-5">
        <UserAvatar user={user} className="w-20 h-20 text-3xl" />
        <div className="space-y-2">
          <div className="flex flex-wrap gap-2">
            <button
              type="button"
              disabled={upload.isPending}
              onClick={() => input.current?.click()}
              className={`${smallButton} bg-slate-900 text-white hover:bg-slate-800 inline-flex items-center gap-1.5`}
            >
              <Camera className="w-4 h-4" />
              {upload.isPending ? 'Загрузка...' : 'Загрузить фото'}
            </button>
            {user.avatarUrl && (
              <button type="button" disabled={remove.isPending} onClick={() => remove.mutate()} className={`${smallButton} bg-slate-100 hover:bg-slate-200`}>
                Удалить
              </button>
            )}
          </div>
          <p className="text-[11px] text-slate-500">JPG, PNG или WebP, до {AVATAR_MAX_MB} МБ</p>
        </div>
        <input
          ref={input}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="sr-only"
          aria-label="Файл аватара"
          onChange={(e) => {
            const file = e.target.files?.[0];
            e.target.value = '';
            if (!file) return;
            setError(null);
            if (file.size > AVATAR_MAX_MB * 1024 * 1024) setError(new Error(`Файл больше ${AVATAR_MAX_MB} МБ`));
            else upload.mutate(file);
          }}
        />
      </div>
      {error ? <ErrorMessage error={error} /> : null}
    </section>
  );
}

function NameCard({ user }: { user: User }) {
  const { setUser } = useSession();
  const [name, setName] = useState(user.name);
  const save = useMutation({ mutationFn: () => sessionApi.updateName(name), onSuccess: setUser });
  const changed = name.trim() !== user.name;

  return (
    <form
      className={card}
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate();
      }}
    >
      <h2 className="text-base font-bold text-slate-900">Данные профиля</h2>
      <Field label="Имя (видно модераторам и в подписи к вашим материалам)" required>
        <Input
          required
          maxLength={100}
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            save.reset();
          }}
        />
      </Field>
      <dl className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-sm">
        <div>
          <dt className="text-xs text-slate-400">Почта</dt>
          <dd className="font-semibold text-slate-800 break-all">{user.email}</dd>
        </div>
        <div>
          <dt className="text-xs text-slate-400">Роль</dt>
          <dd className="font-semibold text-slate-800">{ROLE_LABELS[user.role]}</dd>
        </div>
        <div>
          <dt className="text-xs text-slate-400">На сайте с</dt>
          <dd className="font-semibold text-slate-800">{formatDateTime(user.createdAt)}</dd>
        </div>
      </dl>
      {save.error && <ErrorMessage error={save.error} />}
      <div className="flex items-center gap-3">
        <button type="submit" disabled={!changed || save.isPending} className={`${smallButton} bg-slate-900 text-white hover:bg-slate-800`}>
          Сохранить
        </button>
        <Saved show={save.isSuccess} />
      </div>
    </form>
  );
}

function PasswordCard() {
  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [repeat, setRepeat] = useState('');
  const save = useMutation({
    mutationFn: () => {
      if (next !== repeat) throw new Error('Новый пароль и повтор не совпадают');
      return sessionApi.changePassword(current, next);
    },
    onSuccess: () => {
      setCurrent('');
      setNext('');
      setRepeat('');
    },
  });

  return (
    <form
      className={card}
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate();
      }}
    >
      <h2 className="text-base font-bold text-slate-900">Смена пароля</h2>
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
        <Field label="Текущий пароль" required>
          <Input type="password" required autoComplete="current-password" value={current} onChange={(e) => setCurrent(e.target.value)} />
        </Field>
        <Field label="Новый пароль" required>
          <Input type="password" required minLength={8} maxLength={72} autoComplete="new-password" value={next} onChange={(e) => setNext(e.target.value)} />
        </Field>
        <Field label="Повторите новый" required>
          <Input type="password" required autoComplete="new-password" value={repeat} onChange={(e) => setRepeat(e.target.value)} />
        </Field>
      </div>
      {save.error && <ErrorMessage error={save.error} />}
      <div className="flex items-center gap-3">
        <Button type="submit" variant="secondary" disabled={save.isPending}>
          Сменить пароль
        </Button>
        <Saved show={save.isSuccess} />
      </div>
    </form>
  );
}

export function Profile({ user }: { user: User }) {
  return (
    <div className="space-y-5">
      <AvatarCard user={user} />
      {/* key: после сохранения имени поле начинается с нового значения */}
      <NameCard key={user.name} user={user} />
      <PasswordCard />
    </div>
  );
}
