import { cn } from '@/shared/lib';

export function ObjectTags({ tags, className, tagClassName }: { tags: string[]; className?: string; tagClassName?: string }) {
  return (
    <div className={cn('flex flex-wrap gap-2', className)}>
      {tags.map((tag) => (
        <span key={tag} className={cn('px-2.5 py-1 bg-slate-100 text-slate-700 rounded text-xs font-medium', tagClassName)}>
          #{tag}
        </span>
      ))}
    </div>
  );
}
