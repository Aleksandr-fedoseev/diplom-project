/* eslint-disable @next/next/no-img-element */
import Link from 'next/link';
import { notFound } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import { findStoryBySlug, relatedStories, StoryBlocks } from '@/entities/story';
import { RelatedObjects } from './RelatedObjects';
import { RelatedStories } from './RelatedStories';
import { StoryHeader } from './StoryHeader';
import { StoryMapSection } from './StoryMapSection';
import { StorySources } from './StorySources';

interface StoryPageProps {
  params: Promise<{ slug: string }>;
}

export default async function StoryPage({ params }: StoryPageProps) {
  const { slug } = await params;
  const story = findStoryBySlug(slug);
  if (!story) notFound();

  const related = relatedStories(story);

  return (
    <article className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      <div className="mb-8">
        <Link
          href="/stories"
          className="inline-flex items-center gap-2 text-xs font-bold text-slate-600 hover:text-slate-950 transition-colors uppercase tracking-wider"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Назад к разделу «Истории»</span>
        </Link>
      </div>

      <StoryHeader story={story} />

      <figure className="rounded-3xl overflow-hidden bg-slate-100 border border-slate-200 shadow-xs mb-10">
        <img src={story.coverImage} alt={story.title} className="w-full h-[320px] sm:h-[480px] object-cover" />
        {story.coverCaption && (
          <figcaption className="p-3.5 bg-slate-50 border-t border-slate-200 text-xs text-slate-500 text-center italic">
            {story.coverCaption}
          </figcaption>
        )}
      </figure>

      <StoryBlocks blocks={story.blocks} />

      {story.mapCoordinates && <StoryMapSection place={story.mapCoordinates} objectSlug={story.relatedObjectSlugs?.[0]} />}
      {story.relatedObjectSlugs?.length ? <RelatedObjects slugs={story.relatedObjectSlugs} /> : null}
      <StorySources sources={story.sources} />
      {related.length > 0 && <RelatedStories stories={related} />}
    </article>
  );
}
