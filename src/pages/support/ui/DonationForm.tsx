'use client';

import { useState } from 'react';
import { Button, Chip, Input } from '@/shared/ui';
import { DONATION_AMOUNTS, donationSum } from '../model/donation';

// Разовое пожертвование. Оплата пока не подключена: кнопка только подтверждает сумму.
export function DonationForm() {
  const [preset, setPreset] = useState<string>(DONATION_AMOUNTS[2]);
  const [custom, setCustom] = useState('');
  const sum = donationSum(preset, custom);

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-2xl font-black text-slate-950">Разовое пожертвование на исследовательские работы</h2>
        <p className="text-xs sm:text-sm text-slate-600 leading-relaxed">
          Все собранные средства направляются на оплату архивных копий в ГАОО, закупку защитных составов для дерева
          и выезды обмерных экспедиций.
        </p>
      </div>

      <div className="space-y-3">
        <span className="block text-xs font-bold uppercase tracking-wider text-slate-700">
          Выберите сумму пожертвования:
        </span>
        <div className="grid grid-cols-3 sm:grid-cols-4 gap-3">
          {DONATION_AMOUNTS.map((amount) => (
            <Chip
              key={amount}
              active={preset === amount && !custom}
              onClick={() => {
                setPreset(amount);
                setCustom('');
              }}
              variant="outline"
              className="py-3 px-4 text-sm"
            >
              {amount} ₽
            </Chip>
          ))}
        </div>

        <div className="pt-2">
          <Input
            type="number"
            min={1}
            placeholder="Другая сумма в рублях..."
            aria-label="Другая сумма в рублях"
            value={custom}
            onChange={(e) => setCustom(e.target.value)}
          />
        </div>
      </div>

      <Button onClick={() => alert(`Спасибо за поддержку! Сумма ${sum} ₽ принята к перечислению.`)}>
        Поддержать проект ({sum} ₽)
      </Button>
    </div>
  );
}
