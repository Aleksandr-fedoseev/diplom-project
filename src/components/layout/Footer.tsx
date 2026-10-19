import React from 'react';
import Link from 'next/link';
import { Landmark, MapPin, Mail, Phone, Heart } from 'lucide-react';

export default function Footer() {
  return (
    <footer className="bg-slate-950 text-slate-400 text-sm border-t border-slate-900 mt-auto">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12 lg:py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-8 lg:gap-12">
          
          {/* Col 1-2: Brand Identity */}
          <div className="lg:col-span-2 space-y-4">
            <div className="flex flex-col">
              <span className="font-black text-2xl tracking-tight text-white uppercase">
                НАСЛЕДИЕ
              </span>
              <span className="text-xs font-bold text-slate-400 tracking-wide mt-0.5">
                Историко-культурная платформа Астраханской области
              </span>
            </div>

            <p className="text-slate-400 text-xs sm:text-sm leading-relaxed max-w-sm">
              Интерактивная геоплатформа историко-культурного наследия региона: каталог архитектурных памятников, архивные источники ГАОО и цифровой музей деревянного зодчества.
            </p>

            <div className="pt-2 text-xs text-slate-500">
              © {new Date().getFullYear()} Интерактивная платформа наследия Астраханской области
            </div>
          </div>

          {/* Col 3: Sections */}
          <div className="space-y-3">
            <h4 className="font-bold text-white text-xs uppercase tracking-wider">
              Разделы
            </h4>
            <ul className="space-y-2 text-xs sm:text-sm">
              <li>
                <Link href="/stories" className="hover:text-white transition-colors">
                  Истории и расследования
                </Link>
              </li>
              <li>
                <Link href="/map" className="hover:text-white transition-colors">
                  Интерактивная карта
                </Link>
              </li>
              <li>
                <Link href="/projects/wooden-museum" className="hover:text-white transition-colors">
                  Музей деревянной архитектуры
                </Link>
              </li>
              <li>
                <Link href="/tours" className="hover:text-white transition-colors">
                  Экскурсии по городу
                </Link>
              </li>
              <li>
                <Link href="/shop" className="hover:text-white transition-colors">
                  Мерч и издания
                </Link>
              </li>
            </ul>
          </div>

          {/* Col 4: Participation & Support */}
          <div className="space-y-3">
            <h4 className="font-bold text-white text-xs uppercase tracking-wider">
              Участие
            </h4>
            <ul className="space-y-2 text-xs sm:text-sm">
              <li>
                <Link href="/support" className="hover:text-white transition-colors flex items-center gap-1.5">
                  <Heart className="w-3.5 h-3.5 text-rose-500 fill-rose-500" />
                  <span>Помочь проекту</span>
                </Link>
              </li>
              <li>
                <Link href="/support#wooden" className="hover:text-white transition-colors">
                  Передать старый наличник
                </Link>
              </li>
              <li>
                <Link href="/support#archive" className="hover:text-white transition-colors">
                  Передать архивное фото
                </Link>
              </li>
              <li>
                <Link href="/support#story" className="hover:text-white transition-colors">
                  Рассказать историю дома
                </Link>
              </li>
            </ul>
          </div>

          {/* Col 5: Contacts */}
          <div className="space-y-3">
            <h4 className="font-bold text-white text-xs uppercase tracking-wider">
              Обратная связь
            </h4>
            <div className="space-y-3 text-xs sm:text-sm">
              <p className="text-slate-400 leading-relaxed">
                Вопросы, предложения и архивные материалы принимаются через форму связи и сообщество проекта:
              </p>
              <div className="flex gap-2 text-xs font-semibold">
                <a href="https://vk.com" target="_blank" rel="noreferrer" className="px-3 py-1.5 bg-slate-900 hover:bg-slate-800 text-white rounded-lg transition-colors">
                  ВКонтакте
                </a>
                <a href="https://t.me" target="_blank" rel="noreferrer" className="px-3 py-1.5 bg-slate-900 hover:bg-slate-800 text-white rounded-lg transition-colors">
                  Telegram
                </a>
              </div>
              <div className="pt-1">
                <Link href="/contacts" className="text-xs font-bold text-slate-300 hover:text-white transition-colors">
                  Форма обратной связи →
                </Link>
              </div>
            </div>
          </div>

        </div>
      </div>
    </footer>
  );
}
