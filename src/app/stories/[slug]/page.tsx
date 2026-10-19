import React from 'react';
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { 
  ArrowLeft, 
  Clock, 
  User, 
  MapPin, 
  Compass, 
  BookMarked,
  Share2,
  Quote,
  Volume2,
  Video,
  ArrowRight
} from 'lucide-react';
import { STORIES } from '@/data/stories';
import { HERITAGE_OBJECTS } from '@/data/heritage';

interface StoryPageProps {
  params: Promise<{ slug: string }>;
}

export default async function StoryDetailPage({ params }: StoryPageProps) {
  const { slug } = await params;
  const story = STORIES.find((s) => s.slug === slug);

  if (!story) {
    notFound();
  }

  // Связанные объекты карты (ТЗ 5.2)
  const relatedObjects = HERITAGE_OBJECTS.filter((obj) => 
    story.relatedObjectSlugs?.includes(obj.slug)
  );

  // Связанные статьи (ТЗ 5.2)
  const relatedStories = STORIES.filter((s) => 
    story.relatedStorySlugs?.includes(s.slug)
  );

  return (
    <article className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* Кнопка возврата */}
      <div className="mb-8">
        <Link
          href="/stories"
          className="inline-flex items-center gap-2 text-xs font-bold text-slate-600 hover:text-slate-950 transition-colors uppercase tracking-wider"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Назад к разделу «Истории»</span>
        </Link>
      </div>

      {/* 5.2. Шапка статьи: категория, дата, время чтения, заголовок, подзаголовок/лид, автор */}
      <header className="space-y-4 mb-8">
        <div className="flex flex-wrap items-center gap-2 text-xs">
          <span className="px-3 py-1 bg-slate-100 text-slate-900 font-bold uppercase tracking-wider rounded-md">
            {story.categoryLabel}
          </span>
          <span className="text-slate-300">•</span>
          <span className="text-slate-500 font-medium">{story.publishedAt}</span>
          <span className="text-slate-300">•</span>
          <span className="text-slate-500 font-medium flex items-center gap-1">
            <Clock className="w-3.5 h-3.5" />
            <span>{story.readTime}</span>
          </span>
        </div>

        {/* Заголовок */}
        <h1 className="text-3xl sm:text-5xl font-black text-slate-950 tracking-tight leading-tight">
          {story.title}
        </h1>

        {/* Подзаголовок / Лид */}
        <p className="text-base sm:text-xl text-slate-600 font-normal leading-relaxed">
          {story.lead}
        </p>

        {/* Автор */}
        <div className="flex items-center justify-between pt-4 border-t border-slate-100">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-slate-950 text-white flex items-center justify-center font-bold text-sm">
              <User className="w-5 h-5" />
            </div>
            <div>
              <div className="text-sm font-bold text-slate-950">{story.author.name}</div>
              <div className="text-xs text-slate-500">{story.author.role}</div>
            </div>
          </div>

          <button
            aria-label="Поделиться статьей"
            className="p-2 text-slate-500 hover:text-slate-950 rounded-lg border border-slate-200 transition-colors"
          >
            <Share2 className="w-4 h-4" />
          </button>
        </div>
      </header>

      {/* Главное изображение с подписью к фотографии (ТЗ 5.2) */}
      <div className="rounded-3xl overflow-hidden bg-slate-100 border border-slate-200 shadow-xs mb-10">
        <img
          src={story.coverImage}
          alt={story.title}
          className="w-full h-[320px] sm:h-[480px] object-cover"
        />
        {story.coverCaption && (
          <div className="p-3.5 bg-slate-50 border-t border-slate-200 text-xs text-slate-500 text-center italic">
            {story.coverCaption}
          </div>
        )}
      </div>

      {/* Основной текст и контентные блоки (ТЗ 5.2): абзацы, цитаты, галереи, аудио, видео */}
      <div className="space-y-8 text-base sm:text-lg text-slate-800 leading-relaxed">
        {story.blocks.map((block, idx) => {
          if (block.type === 'paragraph') {
            return (
              <p key={idx} className="leading-relaxed">
                {block.text}
              </p>
            );
          }

          {/* Цитаты */}
          if (block.type === 'quote') {
            return (
              <figure key={idx} className="my-8 p-6 sm:p-8 bg-slate-50 border-l-4 border-slate-950 rounded-r-2xl space-y-2">
                <Quote className="w-8 h-8 text-slate-400 mb-2" />
                <blockquote className="text-lg sm:text-xl font-medium text-slate-900 italic leading-snug">
                  {block.text}
                </blockquote>
                {block.author && (
                  <figcaption className="text-xs sm:text-sm font-semibold text-slate-500 pt-2">
                    — {block.author}
                  </figcaption>
                )}
              </figure>
            );
          }

          {/* Галереи и фотографии с подписями */}
          if (block.type === 'gallery' && block.images) {
            return (
              <div key={idx} className="my-8 grid grid-cols-1 sm:grid-cols-2 gap-4">
                {block.images.map((img, imgIdx) => (
                  <div key={imgIdx} className="rounded-2xl overflow-hidden border border-slate-200 bg-slate-50">
                    <img
                      src={img.url}
                      alt={img.caption}
                      className="w-full h-52 object-cover"
                    />
                    <div className="p-3 text-xs text-slate-600">
                      <strong className="block text-slate-900">{img.caption}</strong>
                      {img.source && <span className="text-slate-400 text-[11px]">{img.source}</span>}
                    </div>
                  </div>
                ))}
              </div>
            );
          }

          {/* Аудио при необходимости (ТЗ 5.2) */}
          if (block.type === 'audio') {
            return (
              <div key={idx} className="my-8 p-5 bg-slate-950 text-white rounded-2xl flex items-center justify-between gap-4">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-10 h-10 rounded-xl bg-white/10 flex items-center justify-center shrink-0">
                    <Volume2 className="w-5 h-5 text-amber-400" />
                  </div>
                  <div className="min-w-0">
                    <div className="text-[10px] text-slate-400 font-bold uppercase tracking-wider">Аудиоматериалы</div>
                    <div className="text-sm font-bold truncate text-white">{block.audioTitle}</div>
                  </div>
                </div>
                <span className="text-xs font-mono font-bold bg-white/10 px-2.5 py-1 rounded text-slate-200 shrink-0">
                  {block.audioDuration}
                </span>
              </div>
            );
          }

          {/* Видео (ТЗ 5.2) */}
          if (block.type === 'video') {
            return (
              <div key={idx} className="my-8 border border-slate-200 rounded-2xl overflow-hidden bg-slate-900">
                <div className="p-4 bg-slate-950 text-white flex items-center gap-2">
                  <Video className="w-4 h-4 text-rose-400" />
                  <span className="text-xs font-bold uppercase tracking-wider">Видеоматериалы: {block.videoTitle}</span>
                </div>
                <div className="relative aspect-video bg-black flex items-center justify-center">
                  <iframe
                    src={block.videoUrl}
                    title={block.videoTitle || 'Видеоматериал'}
                    className="w-full h-full border-0"
                    allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
                    allowFullScreen
                  />
                </div>
              </div>
            );
          }

          return null;
        })}
      </div>

      {/* Карты (ТЗ 5.2) */}
      {story.mapCoordinates && (
        <section className="mt-12 p-6 bg-slate-50 border border-slate-200 rounded-3xl space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-950">
              <Compass className="w-4 h-4 text-blue-600" />
              <span>Карта объекта</span>
            </div>
            <Link
              href={story.relatedObjectSlugs && story.relatedObjectSlugs[0] ? `/map?object=${story.relatedObjectSlugs[0]}` : '/map'}
              className="text-xs font-bold text-slate-900 hover:text-blue-600 flex items-center gap-1 transition-colors"
            >
              <span>Открыть на карте</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>
          <div className="bg-white p-4 rounded-xl border border-slate-200 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <h4 className="text-sm font-bold text-slate-950">{story.mapCoordinates.title}</h4>
              <p className="text-xs text-slate-500 mt-0.5">{story.mapCoordinates.address}</p>
            </div>
            <div className="text-xs font-mono bg-slate-100 px-3 py-1.5 rounded text-slate-700 self-start sm:self-auto">
              {story.mapCoordinates.lat.toFixed(4)}° N, {story.mapCoordinates.lng.toFixed(4)}° E
            </div>
          </div>
        </section>
      )}

      {/* Связанные объекты карты (ТЗ 5.2) */}
      {relatedObjects.length > 0 && (
        <section className="mt-12 p-6 sm:p-8 bg-slate-50 border border-slate-200 rounded-3xl space-y-4">
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-950">
            <MapPin className="w-4 h-4 text-blue-600" />
            <span>Связанные объекты карты</span>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {relatedObjects.map((obj) => (
              <Link
                key={obj.id}
                href={`/map?object=${obj.slug}`}
                className="flex items-center gap-3 p-3.5 bg-white border border-slate-200 rounded-2xl hover:border-slate-400 transition-all group shadow-xs"
              >
                {obj.coverImage ? (
                  <img
                    src={obj.coverImage}
                    alt={obj.title}
                    className="w-16 h-16 rounded-xl object-cover shrink-0"
                  />
                ) : (
                  <div className="w-16 h-16 rounded-xl bg-slate-100 flex items-center justify-center shrink-0">
                    <MapPin className="w-6 h-6 text-slate-400" />
                  </div>
                )}
                <div className="min-w-0 flex-1">
                  <span className="text-[10px] font-bold text-blue-600 uppercase">
                    {obj.yearBuilt} г.
                  </span>
                  <h4 className="text-sm font-bold text-slate-950 group-hover:text-blue-600 transition-colors truncate">
                    {obj.title}
                  </h4>
                  <p className="text-xs text-slate-500 truncate">
                    {obj.address}
                  </p>
                  <span className="text-[11px] font-bold text-slate-900 group-hover:text-blue-600 block mt-1">
                    Открыть на карте →
                  </span>
                </div>
              </Link>
            ))}
          </div>
        </section>
      )}

      {/* Источники (ТЗ 5.2) */}
      <section className="mt-12 p-6 sm:p-8 bg-slate-50 border border-slate-200 border-l-4 border-l-slate-950 rounded-r-3xl space-y-4">
        <div>
          <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-950">
            <BookMarked className="w-4 h-4" />
            <span>Источники</span>
          </div>
          <p className="text-xs text-slate-500 mt-1">
            Материал проверен по архивным фондам ГАОО и верифицированным библиографическим источникам:
          </p>
        </div>

        <div className="space-y-2.5 divide-y divide-slate-200/60 pt-2">
          {story.sources.map((src, i) => (
            <div key={i} className="pt-2.5 first:pt-0 flex items-start gap-3 text-xs sm:text-sm">
              <span className="px-2 py-0.5 bg-white border border-slate-200 rounded text-[10px] font-bold text-slate-700 uppercase shrink-0 mt-0.5">
                {src.typeLabel}
              </span>
              <span className="text-slate-800 font-mono leading-relaxed">
                {src.citation}
              </span>
            </div>
          ))}
        </div>
      </section>

      {/* Связанные статьи (ТЗ 5.2) */}
      {relatedStories.length > 0 && (
        <section className="mt-12 pt-8 border-t border-slate-200 space-y-6">
          <h3 className="text-xl font-bold text-slate-950">
            Связанные статьи
          </h3>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {relatedStories.map((rel) => (
              <Link
                key={rel.id}
                href={`/stories/${rel.slug}`}
                className="p-5 bg-slate-50 border border-slate-200 rounded-2xl hover:border-slate-400 transition-all flex items-center justify-between group"
              >
                <div>
                  <span className="text-[10px] font-bold uppercase tracking-wider text-slate-400">
                    {rel.categoryLabel}
                  </span>
                  <h4 className="text-sm font-bold text-slate-950 group-hover:text-blue-600 transition-colors mt-0.5">
                    {rel.title}
                  </h4>
                  <span className="text-xs text-slate-500 mt-1 block">
                    {rel.readTime}
                  </span>
                </div>
                <span className="inline-flex items-center gap-1 text-xs font-bold text-slate-900 group-hover:text-blue-600 shrink-0 ml-3">
                  <span>Читать статью</span>
                  <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
                </span>
              </Link>
            ))}
          </div>
        </section>
      )}

    </article>
  );
}
