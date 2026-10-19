'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { 
  Search, 
  Menu, 
  X, 
  User, 
  Heart,
  ExternalLink,
  ChevronDown
} from 'lucide-react';

export default function Header() {
  const [isMobileOpen, setIsMobileOpen] = useState(false);
  const [isSearchOpen, setIsSearchOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const pathname = usePathname();

  const navLinks = [
    { href: '/', label: 'Главная' },
    { href: '/stories', label: 'Истории' },
    { href: '/projects', label: 'Проекты' },
    { href: '/shop', label: 'Мерч' },
    { href: '/tours', label: 'Экскурсии' },
    { href: '/contacts', label: 'Где нас найти' },
  ];

  return (
    <header className="sticky top-0 z-50 bg-white/95 backdrop-blur-md border-b border-slate-200 transition-all">
      {/* Main Navigation Bar */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16 sm:h-20">
          
          {/* Logo */}
          <Link href="/" className="flex items-center gap-3 group">
            <div className="flex flex-col">
              <span className="font-black text-xl sm:text-2xl tracking-tight text-slate-950 uppercase group-hover:text-slate-800 transition-colors">
                НАСЛЕДИЕ
              </span>
              <span className="text-[10px] sm:text-[11px] font-bold text-slate-500 tracking-wide leading-none mt-0.5">
                Астраханская область
              </span>
            </div>
          </Link>

          {/* Desktop Navigation Links */}
          <nav className="hidden lg:flex items-center gap-1 xl:gap-2">
            {navLinks.map((link) => {
              const isActive = link.href === '/' 
                ? pathname === '/' 
                : pathname?.startsWith(link.href);

              return (
                <Link
                  key={link.href}
                  href={link.href}
                  className={`px-3 py-2 rounded-lg text-sm font-semibold transition-colors ${
                    isActive
                      ? 'text-slate-950 bg-slate-100 font-bold'
                      : 'text-slate-600 hover:text-slate-950 hover:bg-slate-50'
                  }`}
                >
                  {link.label}
                </Link>
              );
            })}
          </nav>

          {/* Right Action Icons & Buttons */}
          <div className="flex items-center gap-2 sm:gap-3">
            
            {/* Search Trigger */}
            <button
              onClick={() => setIsSearchOpen(!isSearchOpen)}
              className="p-2 text-slate-600 hover:text-slate-950 hover:bg-slate-100 rounded-lg transition-colors"
              aria-label="Поиск по сайту"
            >
              <Search className="w-5 h-5" />
            </button>

            {/* Auth / Profile */}
            <Link
              href="/login"
              className="p-2 text-slate-600 hover:text-slate-950 hover:bg-slate-100 rounded-lg transition-colors"
              aria-label="Личный кабинет"
            >
              <User className="w-5 h-5" />
            </Link>

            {/* Button Support */}
            <Link
              href="/support"
              className="hidden sm:inline-flex items-center gap-2 px-4 py-2.5 bg-slate-950 hover:bg-slate-800 text-white rounded-lg text-xs font-bold uppercase tracking-wider transition-colors shadow-xs"
            >
              <Heart className="w-3.5 h-3.5 text-rose-400 fill-rose-400" />
              <span>Поддержать</span>
            </Link>

            {/* Mobile Menu Trigger */}
            <button
              onClick={() => setIsMobileOpen(!isMobileOpen)}
              className="lg:hidden p-2 text-slate-600 hover:text-slate-950 hover:bg-slate-100 rounded-lg transition-colors"
              aria-label="Открыть меню"
            >
              {isMobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>

          </div>

        </div>
      </div>

      {/* Expandable Search Input Bar */}
      {isSearchOpen && (
        <div className="border-t border-slate-200 bg-slate-50 px-4 py-3 animate-in fade-in slide-in-from-top-2 duration-150">
          <div className="max-w-3xl mx-auto relative">
            <Search className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              autoFocus
              placeholder="Поиск по статьям, улицам, особнякам, источникам ГАОО..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full bg-white border border-slate-200 pl-10 pr-10 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-semibold text-slate-400 hover:text-slate-700"
              >
                Очистить
              </button>
            )}
          </div>
        </div>
      )}

      {/* Mobile Drawer */}
      {isMobileOpen && (
        <div className="lg:hidden border-b border-slate-200 bg-white px-4 pt-3 pb-6 space-y-2 animate-in fade-in slide-in-from-top-2 duration-150">
          {navLinks.map((link) => (
            <Link
              key={link.href}
              href={link.href}
              onClick={() => setIsMobileOpen(false)}
              className="flex items-center justify-between px-3 py-2.5 rounded-lg text-slate-800 font-semibold text-base hover:bg-slate-100"
            >
              <span>{link.label}</span>
            </Link>
          ))}
          <div className="pt-4 border-t border-slate-100 flex flex-col gap-2">
            <Link
              href="/support"
              onClick={() => setIsMobileOpen(false)}
              className="w-full py-3 px-4 bg-slate-950 text-white rounded-lg text-sm font-bold uppercase tracking-wider text-center flex items-center justify-center gap-2"
            >
              <Heart className="w-4 h-4 text-rose-400 fill-rose-400" />
              <span>Помочь проекту</span>
            </Link>
          </div>
        </div>
      )}
    </header>
  );
}
