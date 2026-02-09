-- Откат: возврат к Pin(25, Pin.OUT) и Pico без W (неполный откат текстов)

UPDATE lesson_steps SET content = 'В Thonny напиши:
from machine import Pin
import time
led = Pin(25, Pin.OUT)
while True:
    led.value(1)
    time.sleep(0.5)
    led.value(0)
    time.sleep(0.5)
Нажми Run (F5). Светодиод на плате начнёт мигать.'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND title = 'Код';

UPDATE lesson_steps SET content = 'Pin(25, Pin.OUT) — говорим: пин 25 будет выходом. led.value(1) — включить, led.value(0) — выключить. time.sleep(0.5) — подождать полсекунды. while True — повторять бесконечно.'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND title = 'Что делает каждая строка';

UPDATE lesson_materials SET url_or_path = E'flowchart TB\n  P[Raspberry Pi Pico]\n  P --> GP[Пин GP25]\n  GP --> L[Встроенный светодиод]'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND kind = 'mermaid';

UPDATE lesson_materials SET title = 'Pico: пин GP25 и встроенный LED'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND kind = 'mermaid';

UPDATE checklist_items SET title = 'Создал проект Raspberry Pi Pico'
WHERE lesson_id = 'b0000007-0000-0000-0000-000000000007' AND title = 'Создал проект Raspberry Pi Pico W';
