import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import Header from '@/components/layout/Header';
import Footer from '@/components/layout/Footer';

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
    <html lang="ru" className={inter.variable}>
      <body className="min-h-screen flex flex-col bg-slate-50 text-slate-900 antialiased">
        <Header />
        <main className="flex-grow">
          {children}
        </main>
        <Footer />
      </body>
    </html>
  );
}
