import { describe, expect, it } from 'vitest';
import { OBJECT_FIXTURES } from '@/entities/heritage-object/testing';
import { EMPTY_OBJECT_FORM, toObjectForm, toObjectInput } from './objectForm';

describe('форма объекта', () => {
  it('объект → форма → данные для API без потерь', () => {
    const o = OBJECT_FIXTURES[0];
    const input = toObjectInput(toObjectForm(o));
    expect(input).toMatchObject({
      slug: o.slug,
      title: o.title,
      coordinates: o.coordinates,
      tags: o.tags,
      registryNumber: o.registryNumber,
      isPublished: true,
    });
  });

  it('пустые необязательные поля — null, теги без пустых', () => {
    const input = toObjectInput({ ...EMPTY_OBJECT_FORM, slug: 'x', coordinates: '46.3, 48.0', tags: 'Модерн, , Банк ' });
    expect(input.architect).toBeNull();
    expect(input.tags).toEqual(['Модерн', 'Банк']);
  });

  it('без координат не отправляется', () => {
    expect(() => toObjectInput(EMPTY_OBJECT_FORM)).toThrow(/координаты/);
  });
});
