import Link from 'next/link';
import { Heart } from 'lucide-react';
import { SOCIAL_LINKS } from '@/shared/config';

const SECTIONS = [
  { href: '/stories', label: 'Истории и расследования' },
  { href: '/map', label: 'Интерактивная карта' },
  { href: '/projects/wooden-museum', label: 'Музей деревянной архитектуры' },
  { href: '/tours', label: 'Экскурсии по городу' },
  { href: '/shop', label: 'Мерч и издания' },
];

const PARTICIPATION = [
  { href: '/support?tab=wooden', label: 'Передать старый наличник' },
  { href: '/support?tab=archive', label: 'Передать архивное фото' },
  { href: '/support?tab=story', label: 'Рассказать историю дома' },
];

const heading = 'font-bold text-snow text-xs uppercase tracking-wider';
const link = 'hover:text-snow transition-colors';

export default function Footer() {
  return (
    <footer className="bg-ink-950 text-ink-400 text-sm border-t border-ink-900 mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 lg:py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-8 lg:gap-12">
          <div className="lg:col-span-2 space-y-4">
            <div className="flex flex-col">
              <span className="font-black text-2xl tracking-tight text-snow uppercase">НАСЛЕДИЕ</span>
              <span className="text-xs font-bold text-ink-400 tracking-wide mt-0.5">
                Историко-культурная платформа Астраханской области
              </span>
            </div>
            <p className="text-ink-400 text-xs sm:text-sm leading-relaxed max-w-sm">
              Интерактивная геоплатформа историко-культурного наследия региона: каталог архитектурных памятников, архивные
              источники ГАОО и цифровой музей деревянного зодчества.
            </p>
            <div className="pt-2 text-xs text-ink-500">
              © {new Date().getFullYear()} Интерактивная платформа наследия Астраханской области
            </div>
          </div>

          <nav aria-label="Разделы" className="space-y-3">
            <h2 className={heading}>Разделы</h2>
            <ul className="space-y-2 text-xs sm:text-sm">
              {SECTIONS.map((s) => (
                <li key={s.href}>
                  <Link href={s.href} className={link}>
                    {s.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>

          <nav aria-label="Участие" className="space-y-3">
            <h2 className={heading}>Участие</h2>
            <ul className="space-y-2 text-xs sm:text-sm">
              <li>
                <Link href="/support" className={`${link} flex items-center gap-1.5`}>
                  <Heart className="w-3.5 h-3.5 text-rose-500 fill-rose-500" />
                  <span>Помочь проекту</span>
                </Link>
              </li>
              {PARTICIPATION.map((p) => (
                <li key={p.href}>
                  <Link href={p.href} className={link}>
                    {p.label}
                  </Link>
                </li>
              ))}
            </ul>
          </nav>

          <div className="space-y-3">
            <h2 className={heading}>Обратная связь</h2>
            <div className="space-y-3 text-xs sm:text-sm">
              <p className="text-ink-400 leading-relaxed">
                Вопросы, предложения и архивные материалы принимаются через форму связи и сообщество проекта:
              </p>
              <div className="flex gap-2 text-xs font-semibold">
                {SOCIAL_LINKS.map((s) => (
                  <a key={s.href} href={s.href} target="_blank" rel="noreferrer" className="px-3 py-1.5 bg-ink-900 hover:bg-ink-800 text-snow rounded-lg transition-colors">
                    {s.label}
                  </a>
                ))}
              </div>
              <Link href="/contacts" className="inline-block pt-1 text-xs font-bold text-ink-300 hover:text-snow transition-colors">
                Форма обратной связи →
              </Link>
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
