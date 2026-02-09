-- Один курс: только Raspberry Pi Pico, с самого начала. Удаляем всё старое (Arduino и старый Pico).
DELETE FROM user_checklist_progress;
DELETE FROM user_lesson_progress;
DELETE FROM lesson_tags;
DELETE FROM lesson_materials;
DELETE FROM lesson_steps;
DELETE FROM checklist_items;
DELETE FROM lessons;
DELETE FROM modules;

-- Единственный модуль: Робототехника с нуля на Raspberry Pi Pico
INSERT INTO modules (id, title, description, sort_order)
VALUES (
  'a0000001-0000-0000-0000-000000000001',
  'Робототехника с нуля: Raspberry Pi Pico',
  'Всё с нуля — от понятия «микроконтроллер» до первой программы. Дешёвая плата Pico и бесплатная среда Thonny.',
  0
);

-- Урок 0: Что такое робототехника и микроконтроллер
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000001-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Что такое робототехника и микроконтроллер',
  'Самые первые понятия: что мы будем изучать и как устроена «мозг» робота.',
  'theory',
  0
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000001-0000-0000-0000-000000000001', 'Что такое робототехника', 'Робототехника — это когда мы собираем устройства (роботов), которые сами что-то делают по программе: ездят, мигают светом, реагируют на кнопки и датчики. Всё начинается с маленькой платы с «мозгом» — микроконтроллером.', 0),
  (gen_random_uuid(), 'b0000001-0000-0000-0000-000000000001', 'Что такое микроконтроллер', 'Микроконтроллер — это мини-компьютер на одной микросхеме. Он выполняет твою программу и управляет «ножками» (пинами): к ним можно подключить светодиоды, кнопки, моторы. Мы будем использовать плату Raspberry Pi Pico — на ней как раз такой микроконтроллер (RP2040).', 1);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000001-0000-0000-0000-000000000001', 'link', '/schematics/microcontroller.svg', 'Микроконтроллер — мини-компьютер');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000001-0000-0000-0000-000000000001', 'raspberry_pi');

-- Урок 1: Что понадобится
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000002-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Что понадобится для занятий',
  'Список того, что нужно купить и подготовить. Всё недорого.',
  'theory',
  1
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000002-0000-0000-0000-000000000001', 'Плата Raspberry Pi Pico', 'Главное — плата Pico (или Pico W с Wi‑Fi). Она стоит недорого, продаётся в магазинах электроники и на маркетплейстах.', 0),
  (gen_random_uuid(), 'b0000002-0000-0000-0000-000000000001', 'Кабель USB', 'Кабель micro USB — чтобы подключать Pico к компьютеру и заливать программу.', 1),
  (gen_random_uuid(), 'b0000002-0000-0000-0000-000000000001', 'Компьютер', 'Компьютер с Windows, macOS или Linux. На него установим бесплатную программу Thonny.', 2),
  (gen_random_uuid(), 'b0000002-0000-0000-0000-000000000001', 'Для первых опытов', 'Светодиоды (3–5 штук), резисторы 220 Ом (5–10 штук), провода «папа-папа» и макетная плата. Можно купить готовый набор «для Pico».', 3);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000002-0000-0000-0000-000000000001', 'link', '/schematics/what-you-need.svg', 'Что понадобится');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000002-0000-0000-0000-000000000001', 'raspberry_pi');

-- Урок 2: Безопасность
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000003-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Безопасность при работе',
  'Простые правила: как не повредить плату и себя.',
  'theory',
  2
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000003-0000-0000-0000-000000000001', 'Отключай питание', 'Когда подключаешь или отключаешь провода — сначала отключи Pico от USB.', 0),
  (gen_random_uuid(), 'b0000003-0000-0000-0000-000000000001', 'Проверяй схему', 'Перед включением ещё раз проверь: светодиод через резистор, плюс и минус не перепутаны (длинная нога светодиода — к плюсу).', 1),
  (gen_random_uuid(), 'b0000003-0000-0000-0000-000000000001', 'Не короть пины', 'Не соединяй два пина проводом без схемы — можно сжечь плату. Делай только то, что показано в уроке.', 2);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000003-0000-0000-0000-000000000001', 'link', '/schematics/safety.svg', 'Безопасность');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000003-0000-0000-0000-000000000001', 'raspberry_pi');

-- Урок 3: Установка Thonny и прошивка Pico
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000004-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Установка Thonny и прошивка Pico',
  'Ставим программу для написания кода и заливаем на Pico прошивку MicroPython.',
  'theory',
  3
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000004-0000-0000-0000-000000000001', 'Скачай Thonny', 'Зайди на сайт thonny.org и скачай Thonny для своей системы (Windows, Mac или Linux). Установи программу.', 0),
  (gen_random_uuid(), 'b0000004-0000-0000-0000-000000000001', 'Подключи Pico в режиме прошивки', 'Зажми кнопку BOOTSEL на плате Pico и, не отпуская, подключи её к компьютеру по USB. Pico появится как съёмный диск — значит, готова к прошивке.', 1),
  (gen_random_uuid(), 'b0000004-0000-0000-0000-000000000001', 'Залить MicroPython', 'Скачай файл прошивки .uf2 с сайта Raspberry Pi (раздел Pico / MicroPython). Перетащи этот файл на «диск» Pico. Плата перезагрузится — прошивка установлена.', 2),
  (gen_random_uuid(), 'b0000004-0000-0000-0000-000000000001', 'Настрой Thonny', 'Подключи Pico снова (уже без BOOTSEL). В Thonny: Run → Configure interpreter → выбери «MicroPython (Raspberry Pi Pico)» и порт, на котором висит плата. Готово.', 3);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000004-0000-0000-0000-000000000001', 'link', '/schematics/thonny-ide.svg', 'Thonny — среда для MicroPython');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000004-0000-0000-0000-000000000001', 'raspberry_pi');

-- Урок 4: Первая схема — светодиод
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000005-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Первая схема: подключаем светодиод',
  'Собираем простую цепь: Pico, резистор и светодиод. По схеме.',
  'practice',
  4
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000005-0000-0000-0000-000000000001', 'Зачем резистор', 'Светодиод нельзя подключать напрямую к пину — сгорит. Резистор 220 Ом ограничивает ток. Всегда: пин → резистор → длинная нога светодиода (плюс) → короткая нога (минус) → GND.', 0),
  (gen_random_uuid(), 'b0000005-0000-0000-0000-000000000001', 'Встроенный светодиод', 'На Pico есть встроенный светодиод на пине GP25. Можно сначала попробовать программу без внешних проводов — он будет мигать.', 1),
  (gen_random_uuid(), 'b0000005-0000-0000-0000-000000000001', 'Внешний светодиод', 'Подключи внешний светодиод: GP25 (или любой другой GPIO) → резистор 220 Ом → длинная нога LED → короткая нога LED → GND. Смотри схему ниже.', 2);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000005-0000-0000-0000-000000000001', 'link', '/schematics/led-circuit.svg', 'Схема подключения светодиода');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000005-0000-0000-0000-000000000001', 'raspberry_pi');

-- Урок 5: Первая программа — мигание
INSERT INTO lessons (id, module_id, title, description, lesson_type, sort_order)
VALUES (
  'b0000006-0000-0000-0000-000000000001',
  'a0000001-0000-0000-0000-000000000001',
  'Первая программа: мигающий светодиод',
  'Пишем код на MicroPython и загружаем его на Pico.',
  'practice',
  5
);
INSERT INTO lesson_steps (id, lesson_id, title, content, sort_order)
VALUES
  (gen_random_uuid(), 'b0000006-0000-0000-0000-000000000001', 'Код', 'В Thonny напиши:
from machine import Pin
import time
led = Pin(25, Pin.OUT)
while True:
    led.value(1)
    time.sleep(0.5)
    led.value(0)
    time.sleep(0.5)
Нажми Run (F5). Светодиод на плате начнёт мигать.', 0),
  (gen_random_uuid(), 'b0000006-0000-0000-0000-000000000001', 'Что делает каждая строка', 'Pin(25, Pin.OUT) — говорим: пин 25 будет выходом. led.value(1) — включить, led.value(0) — выключить. time.sleep(0.5) — подождать полсекунды. while True — повторять бесконечно.', 1);
INSERT INTO lesson_materials (id, lesson_id, kind, url_or_path, title)
VALUES (gen_random_uuid(), 'b0000006-0000-0000-0000-000000000001', 'link', '/schematics/pico-board.svg', 'Pico: пин GP25 и встроенный LED');
INSERT INTO lesson_tags (lesson_id, tag) VALUES ('b0000006-0000-0000-0000-000000000001', 'raspberry_pi');
