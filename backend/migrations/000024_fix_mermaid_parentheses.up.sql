-- Круглые скобки в подписи Code[Pin(LED, Pin.OUT)] ломают Mermaid (Parse error). Убираем скобки.
UPDATE lesson_materials
SET url_or_path = E'flowchart TB\n  P[Raspberry Pi Pico W]\n  P --> L[Встроенный светодиод LED]\n  L --> Code[Код: Pin LED Pin.OUT]'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001' AND kind = 'mermaid';
