import Link from 'next/link';
import { ArrowRight, Building2, Coins, Heart, MessageSquare, Upload, type LucideIcon } from 'lucide-react';

// iconColor — цвет значка, hoverColor — цвет заголовка при наведении.
const WAYS: { tab: string; icon: LucideIcon; iconColor: string; hoverColor: string; title: string; text: string; action: string }[] = [
  { tab: 'money', icon: Coins, iconColor: 'text-amber-400', hoverColor: 'group-hover:text-amber-300', title: 'Пожертвование', text: 'На закупку составов для консервации дерева и архивные копии.', action: 'Поддержать' },
  { tab: 'wooden', icon: Building2, iconColor: 'text-emerald-400', hoverColor: 'group-hover:text-emerald-300', title: 'Передать деталь', text: 'Старый наличник, дверь или карниз для музея.', action: 'В музей' },
  { tab: 'archive', icon: Upload, iconColor: 'text-blue-400', hoverColor: 'group-hover:text-blue-300', title: 'Архивные фото', text: 'Снимки, открытки и документы из семейных архивов.', action: 'Прислать фото' },
  { tab: 'story', icon: MessageSquare, iconColor: 'text-purple-400', hoverColor: 'group-hover:text-purple-300', title: 'История дома', text: 'Воспоминания о семье или старинном здании.', action: 'Поделиться' },
];

// Как жители могут помочь — каждая карточка открывает нужную вкладку страницы «Поддержать».
export function Participation() {
  return (
    <section className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 w-full">
      <div className="bg-slate-950 text-white rounded-3xl p-8 sm:p-10 border border-slate-900">
        <div className="max-w-2xl space-y-2 mb-8">
          <div className="inline-flex items-center gap-2 px-3 py-1 rounded bg-white/20 backdrop-blur-md text-[11px] font-bold uppercase tracking-wider text-white">
            <Heart className="w-3.5 h-3.5 text-rose-400 fill-rose-400" />
            <span>Участие жителей</span>
          </div>
          <h2 className="text-2xl sm:text-3xl font-black text-white tracking-tight">Помогите сохранить историю города</h2>
          <p className="text-xs sm:text-sm text-slate-300 leading-relaxed">
            Вы можете внести свой вклад в сохранение культурного наследия Астрахани:
          </p>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {WAYS.map(({ tab, icon: Icon, iconColor, hoverColor, title, text, action }) => (
            <Link
              key={tab}
              href={`/support?tab=${tab}`}
              className="p-5 bg-slate-900 hover:bg-slate-850 border border-slate-800 rounded-2xl transition-all group flex flex-col justify-between"
            >
              <div className="space-y-2">
                <Icon className={`w-5 h-5 ${iconColor}`} />
                <h3 className={`text-sm font-bold text-white transition-colors ${hoverColor}`}>{title}</h3>
                <p className="text-xs text-slate-400 leading-relaxed">{text}</p>
              </div>
              <span className="pt-4 text-xs font-bold text-white flex items-center gap-1">
                <span>{action}</span>
                <ArrowRight className="w-3.5 h-3.5 group-hover:translate-x-1 transition-transform" />
              </span>
            </Link>
          ))}
        </div>
      </div>
    </section>
  );
}
