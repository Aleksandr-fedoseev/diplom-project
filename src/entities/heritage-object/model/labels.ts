import type { HeritageCategory, ProtectionStatus } from './types';

export const CATEGORY_LABELS: Record<HeritageCategory, string> = {
  wooden: 'Деревянное зодчество',
  manor: 'Купеческие усадьбы',
  temple: 'Кремль и храмы',
  modern: 'Модерн',
  lost: 'Утраченное',
};

// Статус охраны в трёх длинах: в списке на карте, в карточке, в паспорте объекта.
export const PROTECTION_LABELS: Record<ProtectionStatus, { short: string; medium: string; long: string }> = {
  federal: { short: 'Фед. ОКН', medium: 'Федеральный ОКН', long: 'Памятник федерального значения' },
  regional: { short: 'Регион. ОКН', medium: 'Региональный ОКН', long: 'Региональное наследие' },
  identified: { short: 'Выявл. ОКН', medium: 'Выявленный ОКН', long: 'Выявленный объект наследия' },
  lost: { short: 'Утрачено', medium: 'Утрачено', long: 'Утраченный объект' },
};
