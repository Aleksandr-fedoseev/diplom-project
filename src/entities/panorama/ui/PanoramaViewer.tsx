'use client';

import React, { useEffect, useRef, useState } from 'react';
import { Sparkles } from 'lucide-react';
import type { Panorama } from '../model/types';

// Pannellum (MIT) лежит в public/vendor/pannellum и подгружается один раз, когда нужен.
interface PannellumViewer {
  destroy(): void;
}

declare global {
  interface Window {
    pannellum?: {
      viewer(container: HTMLElement, config: Record<string, unknown>): PannellumViewer;
    };
  }
}

let pannellumLoading: Promise<void> | null = null;

function loadPannellum(): Promise<void> {
  if (window.pannellum) return Promise.resolve();
  if (!pannellumLoading) {
    pannellumLoading = new Promise((resolve, reject) => {
      const css = document.createElement('link');
      css.rel = 'stylesheet';
      css.href = '/vendor/pannellum/pannellum.css';
      document.head.appendChild(css);

      const script = document.createElement('script');
      script.src = '/vendor/pannellum/pannellum.js';
      script.onload = () => resolve();
      script.onerror = () => {
        pannellumLoading = null;
        reject(new Error('Не удалось загрузить Pannellum'));
      };
      document.head.appendChild(script);
    });
  }
  return pannellumLoading;
}

interface PanoramaViewerProps {
  panoramas: Panorama[];
  heightClassName?: string;
}

export default function PanoramaViewer({
  panoramas,
  heightClassName = 'h-[480px] sm:h-[540px]',
}: PanoramaViewerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [index, setIndex] = useState(0);
  const [failed, setFailed] = useState(false);
  const current = panoramas[Math.min(index, panoramas.length - 1)];
  const url = current?.url;

  useEffect(() => {
    if (!url) return;
    let viewer: PannellumViewer | undefined;
    let cancelled = false;

    loadPannellum()
      .then(() => {
        if (cancelled || !containerRef.current || !window.pannellum) return;
        viewer = window.pannellum.viewer(containerRef.current, {
          type: 'equirectangular',
          panorama: url,
          autoLoad: true,
          hfov: 100,
          showFullscreenCtrl: true,
          strings: { loadingLabel: 'Загрузка панорамы…' },
        });
      })
      .catch(() => setFailed(true));

    return () => {
      cancelled = true;
      viewer?.destroy();
    };
  }, [url]);

  if (!current) return null;
  const isAI = current.source === 'ai_reconstruction';

  return (
    <div className="space-y-3">
      {panoramas.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {panoramas.map((p, i) => (
            <button
              key={p.url}
              onClick={() => setIndex(i)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold border transition-colors ${
                i === index
                  ? 'bg-blue-600 border-blue-600 text-white'
                  : 'bg-white border-slate-200 text-slate-700 hover:bg-slate-50'
              }`}
            >
              {p.caption}
            </button>
          ))}
        </div>
      )}

      <div className={`relative rounded-2xl overflow-hidden bg-slate-950 border border-slate-800 shadow-xl ${heightClassName}`}>
        <div ref={containerRef} className="absolute inset-0" />

        {isAI && (
          <span className="absolute top-3 right-3 z-10 inline-flex items-center gap-1.5 bg-amber-100 text-amber-900 text-xs font-bold px-2.5 py-1 rounded-lg shadow">
            <Sparkles className="w-3.5 h-3.5" />
            Реконструкция ИИ
          </span>
        )}

        {failed && (
          <div className="absolute inset-0 flex items-center justify-center text-sm text-slate-300 p-6 text-center">
            Не удалось загрузить просмотрщик панорам
          </div>
        )}
      </div>

      <p className="text-xs text-slate-500">Тяните мышью или пальцем, чтобы осмотреться; колесо — приближение.</p>

      {isAI && (
        <div className="p-3 bg-amber-50 border border-amber-200 rounded-lg text-xs text-amber-900">
          Панорама создана нейросетью по фотографии объекта: в центре — настоящий снимок, окружение
          дорисовано и может не совпадать с действительностью.
        </div>
      )}
    </div>
  );
}
