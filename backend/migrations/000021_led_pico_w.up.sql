-- Первая программа: Pin("LED", Pin.OUT) и time.sleep(0.1); в Wokwi — Pico W

-- Урок «Первая программа: мигающий светодиод» (b0000006)
UPDATE lesson_steps SET content = 'В Thonny напиши:
from machine import Pin
import time
time.sleep(0.1) # Wait for USB to become ready

led = Pin("LED", Pin.OUT)
while True:
    led.value(1)
    time.sleep(0.5)
    led.value(0)
    time.sleep(0.5)
Нажми Run (F5). Светодиод на плате начнёт мигать.'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND title = 'Код';

UPDATE lesson_steps SET content = 'Pin("LED", Pin.OUT) — встроенный светодиод на плате (на Pico W он так и обозначается). led.value(1) — включить, led.value(0) — выключить. time.sleep(0.5) — подождать полсекунды. time.sleep(0.1) в начале даёт USB время подготовиться. while True — повторять бесконечно.'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND title = 'Что делает каждая строка';

-- Схема: Pico W и встроенный LED
UPDATE lesson_materials SET url_or_path = E'flowchart TB\n  P[Raspberry Pi Pico W]\n  P --> L[Встроенный светодиод LED]\n  L --> Code[Pin("LED", Pin.OUT)]'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND kind = 'mermaid';

UPDATE lesson_materials SET title = 'Pico W: встроенный светодиод LED'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND kind = 'mermaid';


-- Урок Wokwi (b0000007): везде Pico W и правильный код
UPDATE lesson_steps SET content = 'Wokwi — это сайт, на котором на экране показывается плата Raspberry Pi Pico W и детали (светодиоды, кнопки и т.д.). Ты пишешь код — и симулятор «выполняет» его, как настоящая плата. Можно пробовать без реального Pico.'
WHERE lesson_id = 'b0000007-0000-0000-0000-000000000007' AND title = 'Что такое Wokwi';

UPDATE lesson_steps SET content = 'Открой браузер и зайди на сайт wokwi.com. Нажми «New Project» или выбери «Raspberry Pi Pico W» и «MicroPython». Откроется окно с кодом и картинкой платы.'
WHERE lesson_id = 'b0000007-0000-0000-0000-000000000007' AND title = 'Как открыть Wokwi';

UPDATE lesson_steps SET content = 'Слева — место, где ты пишешь код (как в Thonny). Справа — рисунок платы Pico W. На ней уже есть встроенный светодиод. Когда программа запустится, ты увидишь, как он мигает.'
WHERE lesson_id = 'b0000007-0000-0000-0000-000000000007' AND title = 'Окно кода и плата';

UPDATE lesson_steps SET content = 'Вставь тот же код, что и в уроке 5 (мигающий светодиод): from machine import Pin, import time, time.sleep(0.1), led = Pin("LED", Pin.OUT), цикл while True с led.value(1), time.sleep(0.5), led.value(0), time.sleep(0.5). Нажми кнопку Run (или зелёный треугольник). Светодиод на экране начнёт мигать.'
WHERE lesson_id = 'b0000007-0000-0000-0000-000000000007' AND title = 'Простая программа';

UPDATE checklist_items SET title = 'Создал проект Raspberry Pi Pico W'
WHERE lesson_id = 'b0000007-0000-0000-0000-000000000007' AND title = 'Создал проект Raspberry Pi Pico';


-- В последующих уроках: в контексте Wokwi везде Pico W
UPDATE lesson_steps SET content = REPLACE(content, 'проект Raspberry Pi Pico', 'проект Raspberry Pi Pico W')
WHERE lesson_id IN (
  'b0000008-0000-0000-0000-000000000008',
  'b0000009-0000-0000-0000-000000000009',
  'b0000010-0000-0000-0000-000000000010',
  'b0000011-0000-0000-0000-000000000011',
  'b0000012-0000-0000-0000-000000000012',
  'b0000013-0000-0000-0000-000000000013',
  'b0000014-0000-0000-0000-000000000014',
  'b0000015-0000-0000-0000-000000000015'
) AND content LIKE '%Raspberry Pi Pico%' AND content NOT LIKE '%Pico W%';

UPDATE lesson_steps SET content = REPLACE(content, 'создай проект Pico', 'создай проект Pico W')
WHERE content LIKE '%создай проект Pico%' AND content NOT LIKE '%Pico W%';

UPDATE lesson_steps SET content = REPLACE(content, 'Открой Wokwi, создай проект Raspberry Pi Pico', 'Открой Wokwi, создай проект Raspberry Pi Pico W')
WHERE content LIKE '%Открой Wokwi, создай проект Raspberry Pi Pico%' AND content NOT LIKE '%Pico W%';
