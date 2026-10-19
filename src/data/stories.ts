import { Story } from '@/types/content';

export const STORIES: Story[] = [
  {
    id: 'story-nikolskaya',
    slug: 'nikolskaya-ulitsa-i-kosa',
    title: 'Улица Никольская: архитектурный ансамбль купеческой Астрахани',
    lead: 'История формирования главного финансового квартала города на стрелке Волги и Кутума в конце XIX – начале XX века.',
    category: 'architecture',
    categoryLabel: 'Архитектура',
    publishedAt: 'Май 2024',
    readTime: '4 мин',
    coverImage: '/heritage/azovo-donskoy-bank/cover.jpg',
    coverCaption: 'Здание Азово-Донского коммерческого банка (арх. Ф.И. Лидваль, 1910 г.)',
    isLeadArticle: true,
    isPopular: true,
    author: {
      name: 'Редакция платформы',
      role: 'Архитектурный отдел',
    },
    mapCoordinates: {
      lat: 46.3533,
      lng: 48.0289,
      title: 'Улица Никольская и Коса',
      address: 'г. Астрахань, ул. Никольская',
    },
    blocks: [
      {
        type: 'paragraph',
        text: 'Никольская улица соединяет волжские пристани с историческим Белым городом. В начале XX века здесь сформировался банковский и торговый центр Астраханской губернии.',
      },
      {
        type: 'quote',
        text: '«Гранитная облицовка цоколя, строгий серый тон и монументальные маски львов подчеркивали надежность одного из крупнейших банков Российской империи на Каспии».',
        author: 'Из отчета губернского архитектора, 1911 г.',
      },
      {
        type: 'paragraph',
        text: 'Ключевые доминанты улицы — Здание Городской биржи (1910 г., арх. В.Б. Вальдовский-Варганек) на стрелке рек и Азово-Донской коммерческий банк (1910 г., арх. Ф.И. Лидваль) в стиле северного модерна с гранитной облицовкой цоколя.',
      },
      {
        type: 'gallery',
        images: [
          {
            url: '/heritage/birzha/cover.jpg',
            caption: 'Здание Городской биржи на стрелке Волги и Кутума',
            source: 'Фотоархив Центра сохранения архитектуры',
          },
          {
            url: '/heritage/dom-agamzhanovoy/cover.jpg',
            caption: 'Доходный дом Агамжановой, ул. Никольская, 6',
            source: 'Фонд ГАОО',
          },
        ],
      },
      {
        type: 'audio',
        audioTitle: 'Аудиогид: Архитектурные детали северного модерна на Никольской',
        audioDuration: '3:45',
      },
      {
        type: 'video',
        videoTitle: 'Видеоисследование: как строился банковский квартал Астрахани',
        videoUrl: 'https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ',
      },
    ],
    sources: [
      {
        type: 'archive',
        typeLabel: 'Архив',
        citation: 'ГААО. Ф. 32 (Строительное отделение губернского правления), оп. 1, д. 1820.',
      },
      {
        type: 'book',
        typeLabel: 'Книга',
        citation: 'Таркова П. Р. «Архитектурный ансамбль купеческой Астрахани». — Астрахань, 2011.',
      },
    ],
    relatedObjectSlugs: ['birzha', 'azovo-donskoy-bank', 'dom-agamzhanovoy', 'dom-tavrizova'],
    relatedStorySlugs: ['dom-s-butylkami-mify-i-fakty', 'derevyannoe-zodchestvo-astrakhani'],
  },
  {
    id: 'story-bagirov',
    slug: 'dom-s-butylkami-mify-i-fakty',
    title: '«Дом с бутылками»: исследование доходного дома купца Багирова',
    lead: 'Разбор популярной городской легенды об ограждении кровли здания на улице Свердлова на основе архивных чертежей.',
    category: 'myths',
    categoryLabel: 'Городские легенды',
    publishedAt: 'Май 2024',
    readTime: '3 мин',
    coverImage: '/heritage/dom-s-butylkami/cover.jpg',
    coverCaption: 'Доходный дом Багирова, ул. Свердлова, 37',
    isPopular: true,
    author: {
      name: 'Редакция платформы',
      role: 'Архитектурный отдел',
    },
    mapCoordinates: {
      lat: 46.3496,
      lng: 48.0385,
      title: 'Доходный дом Багирова',
      address: 'г. Астрахань, ул. Свердлова, 37',
    },
    blocks: [
      {
        type: 'paragraph',
        text: 'Среди жителей и недобросовестных экскурсоводов распространена байка о том, что парапет дома украшен бутылками из-под шампанского из-за винной торговли купца Багирова.',
      },
      {
        type: 'quote',
        text: '«Никаких бутылок в конструкции никогда не было. По архивным проектным чертежам элементы представляют собой фигурные балясины-пинакли из лекального кирпича — классический прием южнорусской эклектики».',
        author: 'Архивный отчет реставрационной мастерской, 2024 г.',
      },
      {
        type: 'paragraph',
        text: 'По архивным проектным чертежам Городской управы элементы представляют собой фигурные балясины-пинакли из лекального кирпича — декоративный прием южнорусской эклектики начала XX века.',
      },
      {
        type: 'gallery',
        images: [
          {
            url: '/heritage/dom-s-butylkami/detail.jpg',
            caption: 'Архитектурный аттик и кирпичные пинакли ограждения парапета',
            source: 'Фотофиксация Центра, 2024',
          },
        ],
      },
    ],
    sources: [
      {
        type: 'archive',
        typeLabel: 'Архив',
        citation: 'ГААО. Ф. 1 (Астраханская городская управа), оп. 3, дело о строительстве дома Багирова, 1904 г.',
      },
    ],
    relatedObjectSlugs: ['dom-s-butylkami'],
    relatedStorySlugs: ['nikolskaya-ulitsa-i-kosa'],
  },
  {
    id: 'story-wooden-kutum',
    slug: 'derevyannoe-zodchestvo-astrakhani',
    title: 'Деревянная архитектура слободы Селения и Закутумья',
    lead: 'Специфика южнорусского плотницкого мастерства: доставка леса с верховьев Волги на белянах и сохранение резьбы.',
    category: 'wooden',
    categoryLabel: 'Деревянное зодчество',
    publishedAt: 'Апрель 2024',
    readTime: '4 мин',
    coverImage: '/heritage/dom-budagova/cover.jpg',
    coverCaption: 'Элементы деревянной усадьбы Будагова на набережной Кутума',
    author: {
      name: 'Редакция платформы',
      role: 'Отдел реставрации',
    },
    mapCoordinates: {
      lat: 46.3562,
      lng: 48.0475,
      title: 'Усадьба Будагова (слобода Селения)',
      address: 'г. Астрахань, ул. Красная Набережная',
    },
    blocks: [
      {
        type: 'paragraph',
        text: 'Из-за отсутствия местных лесов строевой лес сплавлялся по Волге на гигантских белянах. В районах Селений и Закутумья сложился уникальный массив деревянных усадеб с глухой корабельной и пропильной резьбой.',
      },
      {
        type: 'quote',
        text: '«Астраханский наличник соединил поволжские традиции резьбы с восточными мотивами: в подзорах встречаются мотивы виноградных лоз и солярные знаки».',
        author: 'Экспедиционные записки Центра сохранения архитектуры',
      },
      {
        type: 'gallery',
        images: [
          {
            url: '/heritage/dom-budagova/detail.jpg',
            caption: 'Фрагмент пропильного резного карниза и лопаток',
            source: 'Экспедиция по Закутумью',
          },
        ],
      },
      {
        type: 'audio',
        audioTitle: 'Интервью с мастером-резчиком: спасение исторического дерева',
        audioDuration: '5:12',
      },
    ],
    sources: [
      {
        type: 'archive',
        typeLabel: 'Архив',
        citation: 'Материалы обследований АНО «Центр сохранения архитектуры», 2023–2024 гг.',
      },
    ],
    relatedObjectSlugs: ['dom-tetyushinova', 'dom-budagova'],
    relatedStorySlugs: ['nikolskaya-ulitsa-i-kosa'],
  },
];
