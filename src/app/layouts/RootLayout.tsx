import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import '../styles/globals.css';
import { Header } from '@/widgets/header';
import { Footer } from '@/widgets/footer';
import { THEME_INIT_SCRIPT } from '@/shared/theme';
import { QueryProvider } from '../query/QueryProvider';

const inter = Inter({
  subsets: ['latin', 'cyrillic'],
  weight: ['400', '500', '600', '700'],
  variable: '--font-inter',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Культурное наследие Астраханской области | Официальная геоплатформа',
  description: 'Электронный каталог памятников архитектуры, интерактивная карта, панорамы 360°, архивные материалы и пешеходные маршруты Астрахани. АНО «Центр сохранения архитектуры».',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    // suppressHydrationWarning: класс dark ставит скрипт до React, это ожидаемое расхождение.
    <html lang="ru" className={inter.variable} suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: THEME_INIT_SCRIPT }} />
      </head>
      <body className="min-h-screen flex flex-col bg-slate-50 text-slate-900 antialiased">
        <QueryProvider>
          <Header />
          <main className="flex-grow">{children}</main>
          <Footer />
        </QueryProvider>
      </body>
    </html>
  );
}
