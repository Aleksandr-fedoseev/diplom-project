'use client';

import { useEffect, useState } from 'react';
import { getToken, mediaUrl } from '@/shared/api';
import { cn } from '@/shared/lib';

// Картинка, которую отдают только вошедшему (файлы заявок видят автор и модератор).
// Обычный <img> не передаёт заголовок Authorization, поэтому файл скачивается fetch'ем.
export function AuthImage({ src, alt, className }: { src: string; alt: string; className?: string }) {
  const [url, setUrl] = useState<string | null>(null);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    let objectUrl: string | null = null;
    const ctrl = new AbortController();
    const token = getToken();
    fetch(mediaUrl(src), { headers: token ? { Authorization: `Bearer ${token}` } : {}, signal: ctrl.signal })
      .then((r) => (r.ok ? r.blob() : Promise.reject(new Error(String(r.status)))))
      .then((blob) => {
        objectUrl = URL.createObjectURL(blob);
        setUrl(objectUrl);
      })
      .catch((e) => e.name !== 'AbortError' && setFailed(true));
    return () => {
      ctrl.abort();
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [src]);

  if (failed) return <div className={cn('bg-slate-100 flex items-center justify-center text-xs text-slate-400', className)}>Не загрузилось</div>;
  if (!url) return <div className={cn('bg-slate-100 animate-pulse', className)} />;
  // eslint-disable-next-line @next/next/no-img-element
  return <img src={url} alt={alt} className={cn('object-cover', className)} />;
}
