import type { ReactNode } from 'react';
import { Building2 } from 'lucide-react';
import { cn } from '@/shared/lib';
import type { HeritageObject } from '../model/types';

interface ObjectCoverProps {
  object: Pick<HeritageObject, 'coverImage' | 'title'>;
  className?: string;
  imageClassName?: string;
  iconClassName?: string;
  children?: ReactNode; // значки поверх снимка
}

// Обложка объекта или заглушка со значком здания, если снимка нет.
export function ObjectCover({ object, className, imageClassName, iconClassName = 'w-6 h-6', children }: ObjectCoverProps) {
  if (!object.coverImage) {
    return (
      <div className={cn('bg-slate-50 flex items-center justify-center text-slate-400', className)}>
        <Building2 className={iconClassName} />
      </div>
    );
  }
  return (
    <div className={cn('bg-slate-100 overflow-hidden relative', className)}>
      {/* eslint-disable-next-line @next/next/no-img-element */}
      <img src={object.coverImage} alt={object.title} className={cn('w-full h-full object-cover', imageClassName)} />
      {children}
    </div>
  );
}
