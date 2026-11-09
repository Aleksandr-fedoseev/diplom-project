'use client';

import { useId, useState } from 'react';
import { Upload, X } from 'lucide-react';

interface FileDropProps {
  title: string;
  hint: string; // допустимые форматы и размер
  accept?: string;
  multiple?: boolean;
  files?: File[];
  onFiles?: (files: File[]) => void;
}

// Выбор файлов кнопкой или перетаскиванием. Без onFiles — только оформление (форма ещё
// не отправляет файлы на сервер).
export function FileDrop({ title, hint, accept, multiple, files = [], onFiles }: FileDropProps) {
  const id = useId();
  const [over, setOver] = useState(false);
  const add = (list: FileList | null) => {
    if (!list || !onFiles) return;
    const picked = Array.from(list);
    onFiles(multiple ? [...files, ...picked] : picked.slice(0, 1));
  };

  return (
    <div className="space-y-2">
      <label
        htmlFor={id}
        onDragOver={(e) => {
          e.preventDefault();
          setOver(true);
        }}
        onDragLeave={() => setOver(false)}
        onDrop={(e) => {
          e.preventDefault();
          setOver(false);
          add(e.dataTransfer.files);
        }}
        className={`block p-4 border-2 border-dashed rounded-xl text-center cursor-pointer transition-colors ${
          over ? 'border-blue-500 bg-blue-50' : 'border-slate-300 bg-white hover:bg-slate-50'
        }`}
      >
        <Upload className="w-6 h-6 text-slate-400 mx-auto mb-1" />
        <span className="text-xs font-bold text-slate-700 block">{title}</span>
        <span className="text-[10px] text-slate-400">{hint}</span>
        <input
          id={id}
          type="file"
          className="sr-only"
          accept={accept}
          multiple={multiple}
          disabled={!onFiles}
          onChange={(e) => {
            add(e.target.files);
            e.target.value = ''; // тот же файл можно выбрать снова после удаления
          }}
        />
      </label>
      {files.length > 0 && (
        <ul className="space-y-1">
          {files.map((f, i) => (
            <li key={`${f.name}-${i}`} className="flex items-center justify-between gap-2 px-3 py-1.5 bg-slate-50 rounded-lg text-xs">
              <span className="truncate">{f.name}</span>
              <button
                type="button"
                aria-label={`Убрать ${f.name}`}
                onClick={() => onFiles?.(files.filter((_, j) => j !== i))}
                className="text-slate-400 hover:text-rose-600"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
