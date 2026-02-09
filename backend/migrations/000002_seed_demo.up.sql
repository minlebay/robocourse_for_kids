-- Минимальный контент для проверки: один модуль и один урок
INSERT INTO modules (id, title, description, sort_order)
VALUES ('11111111-1111-1111-1111-111111111101', 'Введение в робототехнику', 'Первый модуль: от светодиода к первому скетчу.', 0)
ON CONFLICT DO NOTHING;

INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  '22222222-2222-2222-2222-222222222201',
  '11111111-1111-1111-1111-111111111101',
  'Мигающий светодиод',
  'Самая простая программа: зажечь и погасить светодиод раз в секунду.',
  'practice',
  0
)
ON CONFLICT DO NOTHING;

INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES (
  gen_random_uuid(),
  '22222222-2222-2222-2222-222222222201',
  'Подключите светодиод',
  'Подключите светодиод к пину 13 и GND через резистор 220 Ом.',
  0
),
(
  gen_random_uuid(),
  '22222222-2222-2222-2222-222222222201',
  'Загрузите скетч',
  'В Arduino IDE откройте пример Blink (Файл → Примеры → 01.Basics → Blink) и загрузите на плату.',
  1
)
ON CONFLICT DO NOTHING;

INSERT INTO lesson_tags (lesson_id, tag)
VALUES ('22222222-2222-2222-2222-222222222201', 'arduino')
ON CONFLICT DO NOTHING;
