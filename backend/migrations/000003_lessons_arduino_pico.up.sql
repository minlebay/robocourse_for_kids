-- Уроки Arduino: что понадобится, установка IDE; модуль Raspberry Pi Pico
-- Сначала сдвигаем порядок существующего урока «Мигающий светодиод»
UPDATE lessons SET sort_order = 2 WHERE id = '22222222-2222-2222-2222-222222222201';

-- Урок: Что понадобится для занятий (Arduino)
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  '33333333-3333-3333-3333-333333333301',
  '11111111-1111-1111-1111-111111111101',
  'Что понадобится для занятий',
  'Список того, что нужно купить и подготовить перед началом работы с Arduino.',
  'theory',
  0
);

INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333301', 'Плата Arduino', 'Плата Arduino Uno или Nano (клоны подойдут — они дешевле). Uno удобнее для начала: большая, хорошо подписанные пины.', 0),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333301', 'Кабель USB', 'Кабель USB type A — mini USB (для Uno) или micro USB (для Nano) для подключения к компьютеру.', 1),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333301', 'Компьютер', 'Компьютер с Windows, macOS или Linux. Дальше понадобится установить Arduino IDE.', 2),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333301', 'Набор для первых уроков', 'Светодиоды (3–5 шт.), резисторы 220 Ом (5–10 шт.), кнопки, провода-папа-папа и макетная плата. Можно купить готовый «стартовый набор Arduino».', 3);

INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333301', 'link', 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/87/Arduino_Logo.svg/320px-Arduino_Logo.svg.png', 'Логотип Arduino');

INSERT INTO lesson_tags (lesson_id, tag) VALUES ('33333333-3333-3333-3333-333333333301', 'arduino');

-- Урок: Установка Arduino IDE
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  '33333333-3333-3333-3333-333333333302',
  '11111111-1111-1111-1111-111111111101',
  'Установка Arduino IDE',
  'Как скачать и установить среду разработки Arduino IDE на компьютер.',
  'theory',
  1
);

INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333302', 'Скачивание', 'Зайди на сайт arduino.cc → Software → Download. Выбери версию для своей ОС (Windows, Mac или Linux) и скачай установщик.', 0),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333302', 'Установка (Windows)', 'Запусти скачанный .exe. Можно оставить галочки по умолчанию. Установи драйверы, если установщик предложит.', 1),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333302', 'Установка (macOS)', 'Открой .dmg и перетащи Arduino IDE в папку «Программы». При первом запуске подтверди открытие приложения из неизвестного разработчика в Настройках.', 2),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333302', 'Первый запуск', 'Подключи плату Arduino к компьютеру по USB. В IDE выбери меню Инструменты → Плата → Arduino Uno (или твоя модель). Порты: Инструменты → Порт — выбери порт, на котором появилась плата.', 3);

INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333302', 'link', 'https://upload.wikimedia.org/wikipedia/commons/thumb/8/87/Arduino_Logo.svg/320px-Arduino_Logo.svg.png', 'Arduino'),
  (gen_random_uuid(), '33333333-3333-3333-3333-333333333302', 'link', 'https://www.arduino.cc/en/software', 'Официальная страница загрузки Arduino IDE');

INSERT INTO lesson_tags (lesson_id, tag) VALUES ('33333333-3333-3333-3333-333333333302', 'arduino');

-- Модуль: Raspberry Pi Pico (дешевле Arduino — микроконтроллер от Raspberry Pi Foundation)
INSERT INTO modules (id, title, description, sort_order)
VALUES (
  '44444444-4444-4444-4444-444444444401',
  'Raspberry Pi Pico: начало',
  'Платформа подешевле Arduino: микроконтроллер Pico, программирование на MicroPython или C. Идеально для второго варианта курса.',
  1
);

-- Урок: Что такое Pico и что понадобится
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  '55555555-5555-5555-5555-555555555501',
  '44444444-4444-4444-4444-444444444401',
  'Что такое Pico и что понадобится',
  'Raspberry Pi Pico — недорогой микроконтроллер. Что купить и чем он отличается от Arduino.',
  'theory',
  0
);

INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555501', 'Что такое Pico', 'Raspberry Pi Pico — это маленькая плата с микроконтроллером RP2040. Она дешевле многих плат Arduino и хорошо подходит для обучения. Программировать можно на MicroPython (проще) или на C.', 0),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555501', 'Что купить', 'Плата Raspberry Pi Pico (или Pico W с Wi‑Fi). Кабель micro USB для подключения к компьютеру. Для первых опытов: светодиоды, резисторы, провода и макетная плата.', 1),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555501', 'Чем отличается от Arduino', 'Другая среда разработки (Thonny для MicroPython), другой язык по желанию (Python вместо C++). Плата меньше и дешевле. Идеально как второй путь: сначала Arduino, потом Pico — или наоборот.', 2);

INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555501', 'link', 'https://upload.wikimedia.org/wikipedia/commons/thumb/2/26/Raspberry_Pi_Pico_%284%29.jpg/320px-Raspberry_Pi_Pico_%284%29.jpg', 'Raspberry Pi Pico');

INSERT INTO lesson_tags (lesson_id, tag) VALUES ('55555555-5555-5555-5555-555555555501', 'raspberry_pi');

-- Урок: Установка Thonny и прошивка Pico
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  '55555555-5555-5555-5555-555555555502',
  '44444444-4444-4444-4444-444444444401',
  'Установка Thonny и прошивка Pico',
  'Устанавливаем среду Thonny и заливаем на Pico прошивку MicroPython.',
  'theory',
  1
);

INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555502', 'Скачивание Thonny', 'Thonny — среда для Python и MicroPython. Сайт: thonny.org. Скачай установщик для своей ОС и установи программу.', 0),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555502', 'Подключение Pico', 'Зажми кнопку BOOTSEL на Pico и подключи плату к компьютеру по USB. Pico появится как съёмный диск — так мы сможем залить прошивку.', 1),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555502', 'Прошивка MicroPython', 'Скачай прошивку MicroPython для Pico с официального сайта (raspberrypi.com/documentation/microcontrollers). Перетащи файл .uf2 на диск Pico. Плата перезагрузится и будет готова к программированию в Thonny.', 2),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555502', 'Настройка Thonny', 'В Thonny: Run → Configure interpreter → выбери «MicroPython (Raspberry Pi Pico)» и порт, на котором висит Pico. Готово — можно писать код.', 3);

INSERT INTO lesson_tags (lesson_id, tag) VALUES ('55555555-5555-5555-5555-555555555502', 'raspberry_pi');

-- Урок: Первый скетч на Pico — светодиод
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  '55555555-5555-5555-5555-555555555503',
  '44444444-4444-4444-4444-444444444401',
  'Первый скетч: светодиод',
  'Подключаем светодиод к Pico и мигаем им из MicroPython.',
  'practice',
  2
);

INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555503', 'Схема', 'Подключи светодиод: длинная нога (плюс) через резистор 220 Ом к GP25 (встроенный светодиод на Pico). Короткая нога — к GND. Либо используй внешний светодиод на любом GPIO и GND.', 0),
  (gen_random_uuid(), '55555555-5555-5555-5555-555555555503', 'Код', 'В Thonny напиши: from machine import Pin\nled = Pin(25, Pin.OUT)\nled.value(1)  # вкл\nled.value(0)  # выкл\nИли в цикле с time.sleep(0.5) — получится мигание.', 1);

INSERT INTO lesson_tags (lesson_id, tag) VALUES ('55555555-5555-5555-5555-555555555503', 'raspberry_pi');
