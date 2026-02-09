-- Урок: Тренажёр — мигающий светодиод без платы
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000007-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Тренажёр: мигающий светодиод без платы',
  'Пока нет Pico под рукой? Запусти симулятор в браузере и попробуй тот же код, что и на реальной плате.',
  'practice',
  6
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000007-0000-0000-0000-000000000001', 'Зачем тренажёр', 'Если у тебя пока нет платы и деталей — не страшно. Ниже есть тренажёр: та же программа на MicroPython, но светодиод нарисован на экране. Напиши код (или отредактируй готовый) и нажми «Запуск».', 0),
  (gen_random_uuid(), 'b0000007-0000-0000-0000-000000000001', 'Как пользоваться', 'В поле для кода можно менять числа: например, в time.sleep(0.5) поставь 1.0 — светодиод будет гореть и не гореть по секунде. Когда достанется настоящая Pico — код будет такой же.', 1);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000007-0000-0000-0000-000000000001', 'simulator', 'led-blink', 'Тренажёр: мигающий светодиод');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000007-0000-0000-0000-000000000001', 'raspberry_pi');

INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000006-0000-0000-0000-000000000001', 'simulator', 'led-blink', 'Нет платы? Попробуй в тренажёре');
