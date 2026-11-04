'use client';

import { useState } from 'react';
import { Heart } from 'lucide-react';
import { Chip } from '@/shared/ui';
import type { MaterialKind } from '../model/materials';
import { ContributionForm } from './ContributionForm';
import { DonationForm } from './DonationForm';
import { OtherWaysToHelp } from './OtherWaysToHelp';

type SupportTab = 'money' | MaterialKind;

const TABS: { id: SupportTab; label: string }[] = [
  { id: 'money', label: '1. Финансовая поддержка' },
  { id: 'wooden', label: '2. Передать элемент резьбы' },
  { id: 'archive', label: '3. Передать архивные фото' },
  { id: 'story', label: '4. Рассказать историю дома' },
];

export default function SupportPage() {
  const [activeTab, setActiveTab] = useState<SupportTab>('money');

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      <div className="max-w-3xl mb-12 space-y-3">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-rose-50 border border-rose-200 text-rose-800 rounded-md text-xs font-bold uppercase tracking-wider">
          <Heart className="w-3.5 h-3.5 text-rose-500 fill-rose-500" />
          <span>Участие в сохранении наследия</span>
        </div>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950">
          Помогите сохранить историю города
        </h1>
        <p className="text-base sm:text-lg text-slate-600 leading-relaxed">
          Поддержать проект можно не только деньгами. Передайте сохранившийся деревянный наличник со сносимого дома,
          поделитесь архивной фотографией из семейного альбома или расскажите историю своего двора.
        </p>
      </div>

      {/* Способы поддержки (ТЗ 12) */}
      <div className="flex items-center gap-2 overflow-x-auto pb-4 mb-10 border-b border-slate-200 scrollbar-none">
        {TABS.map((tab) => (
          <Chip
            key={tab.id}
            active={activeTab === tab.id}
            onClick={() => setActiveTab(tab.id)}
            className="px-4 py-3 text-xs uppercase tracking-wider whitespace-nowrap"
          >
            {tab.label}
          </Chip>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-start">
        <div className="lg:col-span-8 bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-10">
          {/* key: при смене вкладки форма начинается заново */}
          {activeTab === 'money' ? <DonationForm /> : <ContributionForm key={activeTab} kind={activeTab} />}
        </div>
        <div className="lg:col-span-4 space-y-6">
          <OtherWaysToHelp />
        </div>
      </div>
    </div>
  );
}
