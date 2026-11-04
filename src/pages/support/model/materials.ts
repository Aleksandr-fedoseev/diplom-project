// Что можно передать Центру и как выглядит форма для каждого вида материала.
export type MaterialKind = 'wooden' | 'archive' | 'story';

interface TextField {
  label: string;
  placeholder: string;
}

export interface MaterialForm {
  title: string;
  intro: string;
  name: TextField;
  contact: TextField;
  address?: TextField; // нужен, если материал привязан к месту
  description: TextField & { rows: number };
  file?: { title: string; hint: string };
  submit: string;
}

export const MATERIAL_FORMS: Record<MaterialKind, MaterialForm> = {
  wooden: {
    title: 'Передать наличник, дверь или элемент деревянного дома',
    intro:
      'У вас сохранился старый резной наличник, дверь или фрагмент карниза? Не выбрасывайте его. Мы бережно сохраним его в Музее деревянной архитектуры.',
    name: { label: 'Ваше имя', placeholder: 'Иван' },
    contact: { label: 'Телефон или Telegram для связи', placeholder: '+7 (999) 000-00-00' },
    address: {
      label: 'Адрес происхождения дома в Астрахани',
      placeholder: 'г. Астрахань, ул. Коммунистическая / Чехова / Селения...',
    },
    description: {
      label: 'Описание предмета и его состояние',
      placeholder:
        'Например: резной наличник со сносимого флигеля, дерево крепкое, требуется демонтаж или уже снят...',
      rows: 4,
    },
    file: { title: 'Прикрепить фото предмета', hint: 'JPG, PNG до 20 МБ' },
    submit: 'Передать элемент в коллекцию музея',
  },
  archive: {
    title: 'Передать архивные материалы',
    intro:
      'Поделитесь старыми фото, письмами, открытками, чертежами или газетными вырезками. Мы бережно отсканируем их с указанием вашего авторства.',
    name: { label: 'Ваше имя', placeholder: 'Елена' },
    contact: { label: 'Контактный телефон или Email', placeholder: 'Ваш e-mail' },
    description: {
      label: 'Описание материалов (период, кто запечатлен, место)',
      placeholder:
        'Опишите, что это за снимок: дореволюционный вид набережной, семейное фото на фоне утраченного здания...',
      rows: 4,
    },
    file: { title: 'Загрузить сканы или фото', hint: 'JPG, PNG, PDF до 30 МБ' },
    submit: 'Отправить материалы в архив',
  },
  story: {
    title: 'Рассказать личную историю дома или улицы',
    intro:
      '«Знаете историю своего дома, улицы или семьи? Расскажите нам». Редакция подготовит проверенный очерк на основе ваших воспоминаний.',
    name: { label: 'Ваше имя', placeholder: 'Михаил' },
    contact: { label: 'Контакт для связи', placeholder: '+7 (999) 000-00-00' },
    address: {
      label: 'Улица, дом или район Астрахани',
      placeholder: 'Например: ул. Никольская / дом купца Губина / район Косы',
    },
    description: {
      label: 'Текст воспоминаний и семейных историй',
      placeholder:
        'Расскажите, кто жил в этом доме, какие события происходили, сохранились ли интересные детали интерьера...',
      rows: 5,
    },
    submit: 'Отправить историю в редакцию',
  },
};
