'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { 
  Heart, 
  Upload, 
  Building2, 
  FileText, 
  MessageSquare, 
  ShoppingBag, 
  Compass, 
  CheckCircle2, 
  ArrowRight,
  ShieldCheck,
  Hammer
} from 'lucide-react';

export default function SupportPage() {
  const [activeTab, setActiveTab] = useState<'money' | 'wooden' | 'archive' | 'story'>('money');
  const [isSubmitted, setIsSubmitted] = useState(false);

  // Donation state
  const [donationAmount, setDonationAmount] = useState('1000');
  const [customAmount, setCustomAmount] = useState('');

  // Form states
  const [name, setName] = useState('');
  const [contact, setContact] = useState('');
  const [address, setAddress] = useState('');
  const [description, setDescription] = useState('');

  const handleFormSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitted(true);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-16 w-full bg-white text-slate-900">
      
      {/* Header */}
      <div className="max-w-3xl mb-12 space-y-3">
        <div className="inline-flex items-center gap-2 px-3 py-1 bg-rose-50 border border-rose-200 text-rose-800 rounded-md text-xs font-bold uppercase tracking-wider">
          <Heart className="w-3.5 h-3.5 text-rose-500 fill-rose-500" />
          <span>Участие в сохранении наследия</span>
        </div>
        <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-slate-950">
          Помогите сохранить историю города
        </h1>
        <p className="text-base sm:text-lg text-slate-600 leading-relaxed">
          Поддержать проект можно не только деньгами. Передайте сохранившийся деревянный наличник со сносимого дома, поделитесь архивной фотографией из семейного альбома или расскажите историю своего двора.
        </p>
      </div>

      {/* Tabs of Support Types (ТЗ 12) */}
      <div className="flex items-center gap-2 overflow-x-auto pb-4 mb-10 border-b border-slate-200 scrollbar-none">
        {[
          { id: 'money', label: '1. Финансовая поддержка' },
          { id: 'wooden', label: '2. Передать элемент резьбы' },
          { id: 'archive', label: '3. Передать архивные фото' },
          { id: 'story', label: '4. Рассказать историю дома' },
        ].map((tab) => (
          <button
            key={tab.id}
            onClick={() => {
              setActiveTab(tab.id as any);
              setIsSubmitted(false);
            }}
            className={`px-4 py-3 rounded-xl text-xs font-bold uppercase tracking-wider whitespace-nowrap transition-all ${
              activeTab === tab.id
                ? 'bg-slate-950 text-white shadow-xs'
                : 'bg-slate-100 text-slate-700 hover:bg-slate-200'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Main Content Area */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 lg:gap-12 items-start">
        
        {/* Form / Action Box (Col 1-8) */}
        <div className="lg:col-span-8 bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-10">
          
          {isSubmitted ? (
            <div className="py-12 text-center space-y-4">
              <CheckCircle2 className="w-14 h-14 text-emerald-600 mx-auto" />
              <h3 className="text-2xl font-black text-slate-950">
                Спасибо за ваш вклад в историю Астрахани!
              </h3>
              <p className="text-sm text-slate-600 max-w-md mx-auto leading-relaxed">
                Мы бережно изучим переданные сведения и свяжемся с вами по контакту: {contact || 'указанному адресу'}.
              </p>
              <button
                onClick={() => setIsSubmitted(false)}
                className="px-6 py-3 bg-slate-950 text-white rounded-xl text-xs font-bold uppercase tracking-wider"
              >
                Отправить еще материал
              </button>
            </div>
          ) : (
            <>
              {/* TAB 1: Money Donation */}
              {activeTab === 'money' && (
                <div className="space-y-6">
                  <div className="space-y-2">
                    <h2 className="text-2xl font-black text-slate-950">
                      Разовое пожертвование на исследовательские работы
                    </h2>
                    <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
                      Все собранные средства направляются на оплату архивных копий в ГАОО, закупку защитных составов для дерева и выезды обмерных экспедиций.
                    </p>
                  </div>

                  <div className="space-y-3">
                    <label className="block text-xs font-bold uppercase tracking-wider text-slate-700">
                      Выберите сумму пожертвования:
                    </label>
                    <div className="grid grid-cols-3 sm:grid-cols-4 gap-3">
                      {['300', '500', '1000', '2500'].map((sum) => (
                        <button
                          key={sum}
                          type="button"
                          onClick={() => {
                            setDonationAmount(sum);
                            setCustomAmount('');
                          }}
                          className={`py-3 px-4 rounded-xl text-sm font-bold border transition-all ${
                            donationAmount === sum && !customAmount
                              ? 'bg-slate-950 text-white border-slate-950 shadow-xs'
                              : 'bg-white text-slate-800 border-slate-200 hover:border-slate-400'
                          }`}
                        >
                          {sum} ₽
                        </button>
                      ))}
                    </div>

                    <div className="pt-2">
                      <input
                        type="number"
                        placeholder="Другая сумма в рублях..."
                        value={customAmount}
                        onChange={(e) => setCustomAmount(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                  </div>

                  <button
                    onClick={() => alert(`Спасибо за поддержку! Сумма ${customAmount || donationAmount} ₽ принята к перечислению.`)}
                    className="w-full py-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider transition-colors shadow-sm"
                  >
                    Поддержать проект ({customAmount || donationAmount} ₽)
                  </button>
                </div>
              )}

              {/* TAB 2: Wooden Artifact */}
              {activeTab === 'wooden' && (
                <form onSubmit={handleFormSubmit} className="space-y-4">
                  <div className="space-y-1 mb-4">
                    <h2 className="text-2xl font-black text-slate-950">
                      Передать наличник, дверь или элемент деревянного дома
                    </h2>
                    <p className="text-xs sm:text-sm text-slate-600">
                      У вас сохранился старый резной наличник, дверь или фрагмент карниза? Не выбрасывайте его. Мы бережно сохраним его в Музее деревянной архитектуры.
                    </p>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                        Ваше имя*
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="Иван"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                        Телефон или Telegram для связи*
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="+7 (999) 000-00-00"
                        value={contact}
                        onChange={(e) => setContact(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                      Адрес происхождения дома в Астрахани*
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="г. Астрахань, ул. Коммунистическая / Чехова / Селения..."
                      value={address}
                      onChange={(e) => setAddress(e.target.value)}
                      className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                      Описание предмета и его состояние*
                    </label>
                    <textarea
                      rows={4}
                      required
                      placeholder="Например: резной наличник со сносимого флигеля, дерево крепкое, требуется демонтаж или уже снят..."
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                    />
                  </div>

                  <div className="p-4 border-2 border-dashed border-slate-300 rounded-xl text-center bg-white cursor-pointer hover:bg-slate-50 transition-colors">
                    <Upload className="w-6 h-6 text-slate-400 mx-auto mb-1" />
                    <span className="text-xs font-bold text-slate-700 block">Прикрепить фото предмета</span>
                    <span className="text-[10px] text-slate-400">JPG, PNG до 20 МБ</span>
                  </div>

                  <button
                    type="submit"
                    className="w-full py-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider transition-colors shadow-sm"
                  >
                    Передать элемент в коллекцию музея
                  </button>
                </form>
              )}

              {/* TAB 3: Archive Materials */}
              {activeTab === 'archive' && (
                <form onSubmit={handleFormSubmit} className="space-y-4">
                  <div className="space-y-1 mb-4">
                    <h2 className="text-2xl font-black text-slate-950">
                      Передать архивные материалы
                    </h2>
                    <p className="text-xs sm:text-sm text-slate-600">
                      Поделитесь старыми фото, письмами, открытками, чертежами или газетными вырезками. Мы бережно отсканируем их с указанием вашего авторства.
                    </p>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                        Ваше имя*
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="Елена"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                        Контактный телефон или Email*
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="Ваш e-mail"
                        value={contact}
                        onChange={(e) => setContact(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                      Описание материалов (период, кто запечатлен, место)*
                    </label>
                    <textarea
                      rows={4}
                      required
                      placeholder="Опишите, что это за снимок: дореволюционный вид набережной, семейное фото на фоне утраченного здания..."
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                    />
                  </div>

                  <div className="p-4 border-2 border-dashed border-slate-300 rounded-xl text-center bg-white cursor-pointer hover:bg-slate-50 transition-colors">
                    <Upload className="w-6 h-6 text-slate-400 mx-auto mb-1" />
                    <span className="text-xs font-bold text-slate-700 block">Загрузить сканы или фото</span>
                    <span className="text-[10px] text-slate-400">JPG, PNG, PDF до 30 МБ</span>
                  </div>

                  <button
                    type="submit"
                    className="w-full py-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider transition-colors shadow-sm"
                  >
                    Отправить материалы в архив
                  </button>
                </form>
              )}

              {/* TAB 4: Personal Story */}
              {activeTab === 'story' && (
                <form onSubmit={handleFormSubmit} className="space-y-4">
                  <div className="space-y-1 mb-4">
                    <h2 className="text-2xl font-black text-slate-950">
                      Рассказать личную историю дома или улицы
                    </h2>
                    <p className="text-xs sm:text-sm text-slate-600">
                      «Знаете историю своего дома, улицы или семьи? Расскажите нам». Редакция подготовит проверенный очерк на основе ваших воспоминаний.
                    </p>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                        Ваше имя*
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="Михаил"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                        Контакт для связи*
                      </label>
                      <input
                        type="text"
                        required
                        placeholder="+7 (999) 000-00-00"
                        value={contact}
                        onChange={(e) => setContact(e.target.value)}
                        className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                      Улица, дом или район Астрахани*
                    </label>
                    <input
                      type="text"
                      required
                      placeholder="Например: ул. Никольская / дом купца Губина / район Косы"
                      value={address}
                      onChange={(e) => setAddress(e.target.value)}
                      className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
                      Текст воспоминаний и семейных историй*
                    </label>
                    <textarea
                      rows={5}
                      required
                      placeholder="Расскажите, кто жил в этом доме, какие события происходили, сохранились ли интересные детали интерьера..."
                      value={description}
                      onChange={(e) => setDescription(e.target.value)}
                      className="w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950"
                    />
                  </div>

                  <button
                    type="submit"
                    className="w-full py-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider transition-colors shadow-sm"
                  >
                    Отправить историю в редакцию
                  </button>
                </form>
              )}
            </>
          )}

        </div>

        {/* Other Support Formats (Col 9-12) */}
        <div className="lg:col-span-4 space-y-6">
          <div className="bg-slate-50 border border-slate-200 rounded-3xl p-6 sm:p-8 space-y-4">
            <span className="text-xs font-bold uppercase tracking-wider text-slate-400">
              Еще способы помочь
            </span>
            <h3 className="text-lg font-bold text-slate-950">
              Купите мерч или закажите экскурсию
            </h3>
            <p className="text-xs text-slate-600 leading-relaxed">
              100% прибыли от наших пешеходных экскурсий и мерча направляется на реставрационные обмеры и сохранение наследия.
            </p>

            <div className="pt-2 space-y-2">
              <Link
                href="/shop"
                className="w-full py-2.5 px-3 bg-white hover:bg-slate-100 border border-slate-200 rounded-xl text-xs font-bold text-slate-950 flex items-center justify-between transition-colors"
              >
                <span>Перейти в магазин мерча</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </Link>
              <Link
                href="/tours"
                className="w-full py-2.5 px-3 bg-white hover:bg-slate-100 border border-slate-200 rounded-xl text-xs font-bold text-slate-950 flex items-center justify-between transition-colors"
              >
                <span>Выбрать экскурсию</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </Link>
            </div>
          </div>
        </div>

      </div>

    </div>
  );
}
