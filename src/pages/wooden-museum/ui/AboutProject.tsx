import { Flame, HelpCircle, Layers, Sparkles, Warehouse, type LucideIcon } from 'lucide-react';

// Ответы на ключевые вопросы о проекте (ТЗ 8).
const ANSWERS: { icon: LucideIcon; iconColor: string; question: string; title: string; text: string }[] = [
  {
    icon: HelpCircle,
    iconColor: 'text-blue-600',
    question: 'Зачем создан проект',
    title: 'Спасение подлинных деталей от безвозвратной утраты',
    text: 'Деревянные дома исторических слобод города ветшают, попадают под программы расселения и сносятся. Проект создан для того, чтобы ценнейшие образцы резного декора — наличники, филенчатые двери, карнизы и подзоры — не уходили на свалку, а бережно демонтировались, сохранялись и становились основой будущего музея.',
  },
  {
    icon: Flame,
    iconColor: 'text-amber-600',
    question: 'Почему необходимо сохранять',
    title: 'Уникальность южнорусской плотницкой традиции',
    text: 'В Астрахани нет местных строевых лесов: вековые сосны и лиственницы доставлялись за тысячи километров по Волге на гигантских барках-белянах. Мастера создали неповторимый сплав северной корабельной глухой резьбы с солярными южными знаками и восточным растительным орнаментом, не имеющий аналогов в России.',
  },
  {
    icon: Layers,
    iconColor: 'text-emerald-600',
    question: 'Как проходит сбор',
    title: 'Экспедиционный мониторинг и бережный демонтаж',
    text: 'Команда Центра ведет учет аварийного деревянного фонда. При подготовке дома к демонтажу волонтеры и реставраторы согласуют вывоз элементов, деликатно снимают детали без раскола древесины, проводят фотофиксацию, научные обмеры и составляют исторический паспорт каждого предмета. Также элементы передают неравнодушные жители.',
  },
  {
    icon: Warehouse,
    iconColor: 'text-purple-600',
    question: 'Где хранятся элементы',
    title: 'Специализированное фондохранилище',
    text: 'Все собранные артефакты размещаются в сухом вентилируемом фондохранилище мастерской Центра. Древесина проходит очистку от пыли и копоти, антисептическую противогрибковую обработку и стабилизирующую консервацию для предотвращения растрескивания и защиты от влаги.',
  },
];

export function AboutProject() {
  return (
    <section className="space-y-6 mb-16">
      <h2 className="text-xl sm:text-2xl font-black text-slate-950 tracking-tight">О проекте сохранения деревянного зодчества</h2>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {ANSWERS.map(({ icon: Icon, iconColor, question, title, text }) => (
          <div key={question} className="p-6 sm:p-7 bg-slate-50 border border-slate-200 rounded-2xl space-y-2">
            <div className="flex items-center gap-2.5 text-xs font-bold uppercase tracking-wider text-slate-900">
              <Icon className={`w-4 h-4 ${iconColor}`} />
              <span>{question}</span>
            </div>
            <h3 className="text-base sm:text-lg font-bold text-slate-950">{title}</h3>
            <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">{text}</p>
          </div>
        ))}
      </div>

      <div className="p-6 sm:p-8 bg-slate-950 text-white rounded-2xl space-y-3">
        <div className="flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-slate-300">
          <Sparkles className="w-4 h-4 text-amber-400" />
          <span>Планы на будущее</span>
        </div>
        <h3 className="text-xl sm:text-2xl font-bold text-white tracking-tight">Открытая музейная экспозиция и ремесленная школа</h3>
        <p className="text-xs sm:text-sm text-slate-300 leading-relaxed max-w-3xl">
          Главная цель проекта — открыть постоянное городское музейное пространство под открытым небом, где каждый сможет
          увидеть шедевры деревянной резьбы вблизи. В планах также запуск открытой плотницкой мастерской по обучению
          традиционной резьбе и создание полной 3D-модели каждого спасенного наличника.
        </p>
      </div>
    </section>
  );
}
