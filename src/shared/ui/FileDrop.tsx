import { Upload } from 'lucide-react';

interface FileDropProps {
  title: string;
  hint: string; // допустимые форматы и размер
}

// Место для прикрепления файлов. Пока только оформление: загрузка появится вместе с API заявок.
export function FileDrop({ title, hint }: FileDropProps) {
  return (
    <div className="p-4 border-2 border-dashed border-slate-300 rounded-xl text-center bg-white cursor-pointer hover:bg-slate-50 transition-colors">
      <Upload className="w-6 h-6 text-slate-400 mx-auto mb-1" />
      <span className="text-xs font-bold text-slate-700 block">{title}</span>
      <span className="text-[10px] text-slate-400">{hint}</span>
    </div>
  );
}
