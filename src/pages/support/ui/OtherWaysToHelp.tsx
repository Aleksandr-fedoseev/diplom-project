import Link from 'next/link';
import { ArrowRight } from 'lucide-react';

const LINKS = [
  { href: '/shop', label: 'Перейти в магазин мерча' },
  { href: '/tours', label: 'Выбрать экскурсию' },
];

export function OtherWaysToHelp() {
  return (
    <div className="bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-8 space-y-4">
      <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Еще способы помочь</span>
      <h3 className="text-lg font-bold text-slate-950">Купите мерч или закажите экскурсию</h3>
      <p className="text-xs text-slate-600 leading-relaxed">
        100% прибыли от наших пешеходных экскурсий и мерча направляется на реставрационные обмеры и сохранение
        наследия.
      </p>

      <div className="pt-2 space-y-2">
        {LINKS.map((link) => (
          <Link
            key={link.href}
            href={link.href}
            className="w-full py-2.5 px-3 bg-white hover:bg-slate-100 border border-slate-200 rounded-xl text-xs font-bold text-slate-950 flex items-center justify-between transition-colors"
          >
            <span>{link.label}</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        ))}
      </div>
    </div>
  );
}
