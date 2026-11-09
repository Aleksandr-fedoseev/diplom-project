import { mediaUrl } from '@/shared/api';
import { cn } from '@/shared/lib';
import type { User } from '../model/types';

// Аватар или первая буква имени на цветном круге (цвет постоянный для пользователя).
const COLORS = ['bg-blue-600', 'bg-emerald-600', 'bg-rose-600', 'bg-amber-600', 'bg-slate-700'];

export function UserAvatar({ user, className }: { user: Pick<User, 'id' | 'name' | 'avatarUrl'>; className?: string }) {
  if (user.avatarUrl) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={mediaUrl(user.avatarUrl)} alt={user.name} className={cn('rounded-full object-cover shrink-0', className)} />;
  }
  return (
    <span
      aria-hidden
      className={cn(
        'rounded-full text-snow font-bold flex items-center justify-center shrink-0',
        COLORS[user.id % COLORS.length],
        className,
      )}
    >
      {user.name.trim().slice(0, 1).toUpperCase() || '?'}
    </span>
  );
}
