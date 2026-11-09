import { describe, expect, it } from 'vitest';
import { EMPTY_DRAFT, fileForm, toSubmissionInput } from './draft';

const draft = { ...EMPTY_DRAFT, title: ' Дом Франгулова ', description: 'История', source: 'ГАОО' };

describe('заявка', () => {
  it('без объекта — новый объект, с объектом — дополнение', () => {
    expect(toSubmissionInput(draft)).toMatchObject({ kind: 'new_object', title: 'Дом Франгулова' });
    expect(toSubmissionInput(draft, 'birzha')).toMatchObject({ kind: 'addition', objectSlug: 'birzha' });
  });

  it('архивный снимок уходит с годом, современный — без', () => {
    const file = new File(['x'], 'a.jpg', { type: 'image/jpeg' });
    const archive = fileForm(file, { ...draft, archive: true, yearTaken: '1910' });
    expect(archive.get('kind')).toBe('archive_photo');
    expect(archive.get('yearTaken')).toBe('1910');
    const photo = fileForm(file, { ...draft, yearTaken: '1910' });
    expect(photo.get('kind')).toBe('photo');
    expect(photo.get('yearTaken')).toBeNull();
    expect(photo.get('sourceNote')).toBe('ГАОО');
  });
});

describe('координаты в заявке', () => {
  it('разбираются из текста, пустые не отправляются', () => {
    expect(toSubmissionInput({ ...draft, coordinates: '46.35, 48.03' }).coordinates).toEqual({ lat: 46.35, lng: 48.03 });
    expect(toSubmissionInput(draft).coordinates).toBeUndefined();
  });
});
