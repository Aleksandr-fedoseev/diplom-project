import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { STORIES, StoryCard } from '@/entities/story';
import { HeroStory } from './HeroStory';
import { OwnProjects } from './OwnProjects';
import { Participation } from './Participation';
import { SectionHeading } from './SectionHeading';

export default function HomePage() {
  const lead = STORIES.find((s) => s.isLeadArticle) ?? STORIES[0];
  const recent = STORIES.filter((s) => s.id !== lead.id).slice(0, 3);

  return (
    <div className="flex flex-col gap-12 sm:gap-16 pb-20 bg-white text-slate-900">
      <HeroStory story={lead} />

      <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
        <SectionHeading
          kicker="Краеведческое медиа"
          title="Свежие материалы"
          action={
            <Link href="/stories" className="text-xs sm:text-sm font-bold text-slate-900 hover:text-blue-600 flex items-center gap-1 transition-colors">
              <span>Все истории</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          }
        />
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {recent.map((story) => (
            <StoryCard key={story.id} story={story} compact />
          ))}
        </div>
      </section>

      <OwnProjects />
      <Participation />
    </div>
  );
}
