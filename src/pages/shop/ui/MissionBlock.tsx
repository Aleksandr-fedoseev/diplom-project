const POINTS = [
  { title: 'Развиваем платформу', text: 'Средства идут на фотофиксацию, создание 3D-панорам и оцифровку архитектурных памятников.' },
  { title: 'Поддерживаем мастеров', text: 'Все изделия шьются и отпечатываются на локальных астраханских производствах.' },
  { title: 'Архивные исследования', text: 'Финансируем работу историков и архивистов по изучению первоисточников в фондах ГАОО.' },
  { title: 'Делаем край ближе', text: 'Красивый способ рассказать о подлинной культуре Нижней Волги друзьям и гостям города.' },
];

// «Больше, чем сувениры» — куда идут деньги от продаж (ТЗ 9).
export function MissionBlock() {
  return (
    <section className="bg-slate-50 border border-slate-200 rounded-3xl p-8 sm:p-12 mb-16 space-y-8">
      <div className="max-w-xl">
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">Миссия фонда</span>
        <h2 className="text-2xl sm:text-3xl font-black text-slate-950 mt-1">Больше, чем сувениры</h2>
        <p className="text-xs sm:text-sm text-slate-600 mt-2 leading-relaxed">
          Покупая сувенирную продукцию, вы напрямую инвестируете в сохранение исторической памяти родного края.
        </p>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
        {POINTS.map((p, i) => (
          <div key={p.title} className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs space-y-2">
            <div className="w-8 h-8 rounded-lg bg-slate-100 flex items-center justify-center font-bold text-slate-900 text-sm">
              {String(i + 1).padStart(2, '0')}
            </div>
            <h3 className="text-sm font-bold text-slate-900">{p.title}</h3>
            <p className="text-xs text-slate-500 leading-relaxed">{p.text}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
