import { describe, expect, it } from 'vitest';
import { donationSum } from './donation';

describe('donationSum', () => {
  it('берёт выбранную кнопкой сумму, если своя не введена', () => {
    expect(donationSum('1000', '')).toBe('1000');
  });

  it('своя сумма важнее выбранной кнопкой', () => {
    expect(donationSum('1000', '750')).toBe('750');
  });

  it.each(['0', '-50', 'abc', '   '])('игнорирует неверную свою сумму %j', (custom) => {
    expect(donationSum('500', custom)).toBe('500');
  });
});
