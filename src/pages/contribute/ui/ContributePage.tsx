'use client';

import React, { Suspense, useId, useState } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, ArrowRight, CheckCircle2, Info, LogIn } from 'lucide-react';
import { useObject } from '@/entities/heritage-object';
import { useSession } from '@/entities/session';
import { isStaff } from '@/entities/user';
import { submissionApi, submissionKeys } from '@/entities/submission';
import { parseCoordinates } from '@/shared/lib';
import { ErrorMessage, FileDrop, Loading, PageHeader } from '@/shared/ui';
import { EMPTY_DRAFT, fileForm, MAX_FILE_MB, STEPS, toSubmissionInput, type ObjectDraft } from '../model/draft';
import { backClass, DraftInput, DraftTextarea, Label, nextClass, submitClass } from './fields';
import { Stepper } from './Stepper';

function Contribute() {
  // /contribute?object=birzha — дополнение к существующему объекту.
  const objectSlug = useSearchParams()?.get('object') || undefined;
  const target = useObject(objectSlug);
  const { isAuthed, isLoading, user } = useSession();
  const qc = useQueryClient();
  const [step, setStep] = useState(1);
  const [draft, setDraft] = useState<ObjectDraft>(EMPTY_DRAFT);
  const [files, setFiles] = useState<File[]>([]);
  const [progress, setProgress] = useState('');
  const id = useId();
  const set = (patch: Partial<ObjectDraft>) => setDraft((d) => ({ ...d, ...patch }));

  // Сначала заявка, потом её файлы по одному. Если файл не принят, заявка уже создана —
  // об этом сообщаем, а не делаем вид, что ничего не отправилось.
  const submit = useMutation({
    mutationFn: async () => {
      const sub = await submissionApi.create(toSubmissionInput(draft, objectSlug));
      const failed: string[] = [];
      for (const [i, file] of files.entries()) {
        setProgress(`Загрузка файла ${i + 1} из ${files.length}...`);
        try {
          await submissionApi.uploadMedia(sub.id, fileForm(file, draft));
        } catch {
          failed.push(file.name);
        }
      }
      return { sub, failed };
    },
    onSettled: () => setProgress(''),
    onSuccess: () => qc.invalidateQueries({ queryKey: submissionKeys.mine }),
  });

  if (isLoading) return <Loading />;

  if (!isAuthed) {
    const next = `/contribute${objectSlug ? `?object=${encodeURIComponent(objectSlug)}` : ''}`;
    return (
      <div className="max-w-xl mx-auto px-4 py-16 text-center space-y-4">
        <h1 className="text-2xl font-bold text-slate-900">Войдите, чтобы предложить материал</h1>
        <p className="text-sm text-slate-600">
          Заявка закрепляется за вашим аккаунтом: в личном кабинете видно, на каком она этапе, а о решении модератора
          придёт письмо.
        </p>
        <Link
          href={`/login?next=${encodeURIComponent(next)}`}
          className="inline-flex items-center gap-2 px-6 py-3 bg-slate-900 text-white rounded-lg font-semibold text-sm hover:bg-slate-800"
        >
          <LogIn className="w-4 h-4" />
          <span>Войти или зарегистрироваться</span>
        </Link>
      </div>
    );
  }

  // Сотрудник добавляет объекты сразу в каталог, без заявки самому себе.
  if (isStaff(user?.role)) {
    return (
      <div className="max-w-xl mx-auto px-4 py-16 text-center space-y-4">
        <h1 className="text-2xl font-bold text-slate-900">Заявки подают пользователи</h1>
        <p className="text-sm text-slate-600">Сотрудник добавляет и правит объекты напрямую в разделе «Модерация».</p>
        <Link href="/moderation?tab=objects" className="inline-flex px-6 py-3 bg-slate-900 text-white rounded-lg font-semibold text-sm hover:bg-slate-800">
          Перейти к объектам
        </Link>
      </div>
    );
  }

  if (submit.data) {
    const { failed } = submit.data;
    return (
      <div className="max-w-2xl mx-auto px-4 py-16 text-center space-y-6">
        <div className="w-16 h-16 bg-emerald-100 text-emerald-600 rounded-full flex items-center justify-center mx-auto">
          <CheckCircle2 className="w-10 h-10" />
        </div>
        <h1 className="text-3xl font-bold text-slate-900">Материал успешно отправлен на модерацию</h1>
        <p className="text-slate-600 text-base leading-relaxed">
          Спасибо за ваш вклад в сохранение культурного наследия Астрахани! Модератор проверит данные, а о решении
          придёт письмо на вашу почту.
        </p>
        {failed.length > 0 && (
          <ErrorMessage error={new Error(`Не удалось загрузить: ${failed.join(', ')}. Заявка отправлена без этих файлов.`)} />
        )}
        <div className="pt-4 flex justify-center gap-3">
          <Link
            href="/account?tab=submissions"
            className="inline-flex items-center gap-2 px-6 py-3 bg-slate-900 text-white rounded-lg font-semibold text-sm hover:bg-slate-800 transition-colors"
          >
            Мои заявки
          </Link>
          <Link href="/map" className="inline-flex items-center px-6 py-3 bg-slate-100 rounded-lg font-semibold text-sm hover:bg-slate-200">
            К карте
          </Link>
        </div>
      </div>
    );
  }

  // Каждый шаг — своя форма: браузер проверит обязательные поля шага, прежде чем пустить дальше.
  const onStepSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (step < STEPS.length) setStep(step + 1);
    else submit.mutate();
  };

  const tooBig = files.filter((f) => f.size > MAX_FILE_MB * 1024 * 1024);
  const back = (
    <button type="button" onClick={() => setStep(step - 1)} className={backClass}>
      <ArrowLeft className="w-4 h-4" />
      <span>Назад</span>
    </button>
  );

  return (
    <div className="max-w-3xl mx-auto px-4 sm:px-6 py-10 sm:py-14 w-full">
      <PageHeader
        className="mb-6"
        kicker="Краудсорсинг архива"
        title={target.data ? `Дополнить: ${target.data.title}` : 'Предложить объект или архивный снимок'}
        lead={
          target.data
            ? 'Добавьте фотографии, архивные снимки или уточнения к объекту каталога.'
            : 'Поделитесь информацией об утраченном или малоизвестном памятнике архитектуры.'
        }
      />

      <Stepper step={step} />

      <form key={step} onSubmit={onStepSubmit} className="bg-white rounded-2xl border border-slate-200 p-6 sm:p-8 shadow-xs space-y-4">
        {step === 1 && (
          <>
            <div>
              <Label htmlFor={`${id}-title`}>{objectSlug ? 'Что вы добавляете*' : 'Название объекта (историческое или народное)*'}</Label>
              <DraftInput
                id={`${id}-title`}
                required
                autoFocus
                placeholder={objectSlug ? 'Например: снимок фасада 1910-х годов' : 'Например: Дом купца Франгулова'}
                value={draft.title}
                onChange={(e) => set({ title: e.target.value })}
              />
            </div>
            {!objectSlug && (
              <div>
                <Label htmlFor={`${id}-address`}>Адрес или ориентир в Астрахани*</Label>
                <DraftInput id={`${id}-address`} required placeholder="г. Астрахань, ул. Советская, д. ..." value={draft.address} onChange={(e) => set({ address: e.target.value })} />
              </div>
            )}
            {!objectSlug && (
              <div>
                <Label htmlFor={`${id}-coords`}>Координаты (широта, долгота), если знаете</Label>
                <DraftInput
                  id={`${id}-coords`}
                  placeholder="Например: 46.3538, 48.0285 — можно скопировать с карты"
                  value={draft.coordinates}
                  onChange={(e) => set({ coordinates: e.target.value })}
                  onBlur={(e) =>
                    e.target.setCustomValidity(
                      e.target.value.trim() && !parseCoordinates(e.target.value) ? 'Два числа: широта и долгота, например 46.3538, 48.0285' : '',
                    )
                  }
                />
              </div>
            )}
            <div>
              <Label htmlFor={`${id}-period`}>Примерное время постройки</Label>
              <DraftInput id={`${id}-period`} placeholder="Например: 1894 г. или конец XIX века" value={draft.period} onChange={(e) => set({ period: e.target.value })} />
            </div>
            <div className="pt-4 flex justify-end">
              <button type="submit" className={nextClass}>
                <span>Далее: История</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </>
        )}

        {step === 2 && (
          <>
            <div>
              <Label htmlFor={`${id}-description`}>Исторические сведения и описание*</Label>
              <DraftTextarea id={`${id}-description`} rows={5} required autoFocus placeholder="Опишите историю здания, известных владельцев, особенности архитектуры..." value={draft.description} onChange={(e) => set({ description: e.target.value })} />
            </div>
            <div>
              <Label htmlFor={`${id}-source`}>Источники информации (архив, семейный архив, литература)</Label>
              <DraftInput id={`${id}-source`} placeholder="Например: ГАОО, фонд 12 или устные воспоминания" value={draft.source} onChange={(e) => set({ source: e.target.value })} />
            </div>
            <div className="pt-4 flex justify-between">
              {back}
              <button type="submit" className={nextClass}>
                <span>Далее: Фотографии</span>
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </>
        )}

        {step === 3 && (
          <>
            <FileDrop
              title="Нажмите для выбора фотографий или перетащите их сюда"
              hint={`JPG, PNG, WebP — до ${MAX_FILE_MB} МБ каждый`}
              accept="image/jpeg,image/png,image/webp"
              multiple
              files={files}
              onFiles={setFiles}
            />
            <label className="flex items-center gap-2 text-sm text-slate-700">
              <input type="checkbox" checked={draft.archive} onChange={(e) => set({ archive: e.target.checked })} />
              <span>Это архивные снимки</span>
            </label>
            {draft.archive && (
              <div>
                <Label htmlFor={`${id}-year`}>Год съёмки</Label>
                <DraftInput id={`${id}-year`} placeholder="Например: 1910 или 1910-е" value={draft.yearTaken} onChange={(e) => set({ yearTaken: e.target.value })} />
              </div>
            )}
            {tooBig.length > 0 && <ErrorMessage error={new Error(`Больше ${MAX_FILE_MB} МБ: ${tooBig.map((f) => f.name).join(', ')}`)} />}
            <div className="p-3 bg-slate-50 rounded-lg text-xs text-slate-500 flex items-start gap-2">
              <Info className="w-4 h-4 text-blue-600 shrink-0 mt-0.5" />
              <span>
                Отправляя данные, вы подтверждаете, что имеете право передать снимки, и соглашаетесь на их научную
                проверку и публикацию на платформе с указанием вашего авторства.
              </span>
            </div>
            {submit.error && <ErrorMessage error={submit.error} />}
            <div className="pt-4 flex justify-between">
              {back}
              <button type="submit" disabled={submit.isPending || tooBig.length > 0} className={submitClass}>
                <CheckCircle2 className="w-4 h-4" />
                <span>{submit.isPending ? progress || 'Отправка...' : 'Отправить на модерацию'}</span>
              </button>
            </div>
          </>
        )}
      </form>
    </div>
  );
}

export default function ContributePage() {
  return (
    <Suspense>
      <Contribute />
    </Suspense>
  );
}
