'use client';

import React, { useState } from 'react';
import { CheckCircle2 } from 'lucide-react';
import { Button, Field, FileDrop, Input, Textarea } from '@/shared/ui';
import { MATERIAL_FORMS, type MaterialKind } from '../model/materials';

interface ContributionFormProps {
  kind: MaterialKind;
}

// Форма передачи материала. Отправка пока только показывает благодарность:
// приём материалов появится вместе с API заявок.
export function ContributionForm({ kind }: ContributionFormProps) {
  const form = MATERIAL_FORMS[kind];
  const [name, setName] = useState('');
  const [contact, setContact] = useState('');
  const [address, setAddress] = useState('');
  const [description, setDescription] = useState('');
  const [isSubmitted, setIsSubmitted] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitted(true);
  };

  if (isSubmitted) {
    return (
      <div className="py-12 text-center space-y-4">
        <CheckCircle2 className="w-14 h-14 text-emerald-600 mx-auto" />
        <h3 className="text-2xl font-black text-slate-950">Спасибо за ваш вклад в историю Астрахани!</h3>
        <p className="text-sm text-slate-600 max-w-md mx-auto leading-relaxed">
          Мы бережно изучим переданные сведения и свяжемся с вами по контакту: {contact || 'указанному адресу'}.
        </p>
        <Button variant="secondary" onClick={() => setIsSubmitted(false)}>
          Отправить еще материал
        </Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-1 mb-4">
        <h2 className="text-2xl font-black text-slate-950">{form.title}</h2>
        <p className="text-xs sm:text-sm text-slate-600">{form.intro}</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <Field label={form.name.label} required>
          <Input required placeholder={form.name.placeholder} value={name} onChange={(e) => setName(e.target.value)} />
        </Field>
        <Field label={form.contact.label} required>
          <Input
            required
            placeholder={form.contact.placeholder}
            value={contact}
            onChange={(e) => setContact(e.target.value)}
          />
        </Field>
      </div>

      {form.address && (
        <Field label={form.address.label} required>
          <Input
            required
            placeholder={form.address.placeholder}
            value={address}
            onChange={(e) => setAddress(e.target.value)}
          />
        </Field>
      )}

      <Field label={form.description.label} required>
        <Textarea
          required
          rows={form.description.rows}
          placeholder={form.description.placeholder}
          value={description}
          onChange={(e) => setDescription(e.target.value)}
        />
      </Field>

      {form.file && <FileDrop title={form.file.title} hint={form.file.hint} />}

      <Button type="submit">{form.submit}</Button>
    </form>
  );
}
