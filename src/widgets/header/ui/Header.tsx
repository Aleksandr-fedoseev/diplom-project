'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Heart, Menu, Search, X } from 'lucide-react';
import { cn } from '@/shared/lib';
import { useSession } from '@/entities/session';
import { isActiveLink, NAV_LINKS, staffLinks } from '../model/nav';
import { SiteSearch } from './SiteSearch';
import { ThemeToggle } from './ThemeToggle';
import { UserMenu } from './UserMenu';

const iconButton = 'p-2 text-slate-600 hover:text-slate-950 hover:bg-slate-100 rounded-lg transition-colors';

export default function Header() {
  const pathname = usePathname();
  const [isMobileOpen, setIsMobileOpen] = useState(false);
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const staff = staffLinks(useSession().user?.role);

  // Переход на другую страницу закрывает меню и поиск.
  useEffect(() => {
    setIsMobileOpen(false);
    setIsSearchOpen(false);
  }, [pathname]);

  return (
    <header className="sticky top-0 z-50 bg-white/95 backdrop-blur-md border-b border-slate-200 transition-all">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 sm:h-20">
          <Link href="/" className="flex flex-col group">
            <span className="font-black text-xl sm:text-2xl tracking-tight text-slate-950 uppercase group-hover:text-slate-800 transition-colors">
              НАСЛЕДИЕ
            </span>
            <span className="text-[10px] sm:text-[11px] font-bold text-slate-500 tracking-wide leading-none mt-0.5">
              Астраханская область
            </span>
          </Link>

          <nav aria-label="Основное меню" className="hidden lg:flex items-center gap-1 xl:gap-2">
            {NAV_LINKS.map((link) => {
              const active = isActiveLink(link.href, pathname);
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  aria-current={active ? 'page' : undefined}
                  className={cn(
                    'px-3 py-2 rounded-lg text-sm font-semibold transition-colors',
                    active ? 'text-slate-950 bg-slate-100 font-bold' : 'text-slate-600 hover:text-slate-950 hover:bg-slate-50',
                  )}
                >
                  {link.label}
                </Link>
              );
            })}
            {staff.length > 0 && <span className="w-px h-6 bg-slate-200 mx-1" aria-hidden />}
            {staff.map((link) => {
              const active = isActiveLink(link.href, pathname);
              return (
                <Link
                  key={link.href}
                  href={link.href}
                  aria-current={active ? 'page' : undefined}
                  className={cn(
                    'px-3 py-2 rounded-lg text-sm font-bold transition-colors',
                    active ? 'bg-blue-600 text-white' : 'text-blue-700 bg-blue-50 hover:bg-blue-100',
                  )}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>

          <div className="flex items-center gap-2 sm:gap-3">
            <button
              type="button"
              onClick={() => setIsSearchOpen((v) => !v)}
              aria-expanded={isSearchOpen}
              className={iconButton}
              aria-label="Поиск по сайту"
            >
              <Search className="w-5 h-5" />
            </button>
            <ThemeToggle />
            <UserMenu />
            <Link
              href="/support"
              className="hidden sm:inline-flex items-center gap-2 px-4 py-2.5 bg-slate-950 hover:bg-slate-800 text-white rounded-lg text-xs font-bold uppercase tracking-wider transition-colors shadow-xs"
            >
              <Heart className="w-3.5 h-3.5 text-rose-400 fill-rose-400" />
              <span>Поддержать</span>
            </Link>
            <button
              type="button"
              onClick={() => setIsMobileOpen((v) => !v)}
              aria-expanded={isMobileOpen}
              className={cn(iconButton, 'lg:hidden')}
              aria-label={isMobileOpen ? 'Закрыть меню' : 'Открыть меню'}
            >
              {isMobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {isSearchOpen && <SiteSearch onDone={() => setIsSearchOpen(false)} />}

      {isMobileOpen && (
        <nav aria-label="Меню" className="lg:hidden border-b border-slate-200 bg-white px-4 pt-3 pb-6 space-y-2 animate-in fade-in slide-in-from-top-2 duration-150">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              aria-current={isActiveLink(link.href, pathname) ? 'page' : undefined}
              className="flex items-center justify-between px-3 py-2.5 rounded-lg text-slate-800 font-semibold text-base hover:bg-slate-100"
            >
              {link.label}
            </Link>
          ))}
          {staff.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              aria-current={isActiveLink(link.href, pathname) ? 'page' : undefined}
              className="flex items-center justify-between px-3 py-2.5 rounded-lg text-blue-700 bg-blue-50 font-bold text-base"
            >
              {link.label}
            </Link>
          ))}
          <div className="pt-4 border-t border-slate-100">
            <Link
              href="/support"
              className="w-full py-3 px-4 bg-slate-950 text-white rounded-lg text-sm font-bold uppercase tracking-wider flex items-center justify-center gap-2"
            >
              <Heart className="w-4 h-4 text-rose-400 fill-rose-400" />
              <span>Помочь проекту</span>
            </Link>
          </div>
        </nav>
      )}
    </header>
  );
}
