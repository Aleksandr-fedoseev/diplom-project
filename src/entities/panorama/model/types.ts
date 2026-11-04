// Панорама 360° (равнопромежуточная проекция 2:1) — те же поля, что у медиафайла
// kind=panorama в API бэкенда.
export interface Panorama {
  url: string;
  source: 'photo' | 'ai_reconstruction'; // ai_reconstruction — подписывается «реконструкция ИИ»
  caption: string;
}
