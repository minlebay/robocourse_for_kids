UPDATE lesson_materials
SET kind = 'mermaid',
    url_or_path = E'flowchart TB\n  subgraph m["Микроконтроллер"]\n    A[Мини-компьютер на одной микросхеме]\n    B[Выполняет программу]\n    C[Управляет пинами: светодиоды, кнопки, моторы]\n  end'
WHERE lesson_id = 'b0000001-0000-0000-0000-000000000001'
  AND url_or_path = '/schematics/microcontroller.svg';

UPDATE lesson_materials
SET kind = 'mermaid',
    url_or_path = E'flowchart TB\n  P[Pico]\n  U[USB-кабель]\n  C[Компьютер]\n  L[Светодиоды, резисторы, макетная плата]\n  P --- U\n  U --- C\n  C --- L'
WHERE lesson_id = 'b0000002-0000-0000-0000-000000000001'
  AND url_or_path = '/schematics/what-you-need.svg';

UPDATE lesson_materials
SET kind = 'mermaid',
    url_or_path = E'flowchart LR\n  A[Отключай питание] --> B[Проверяй схему]\n  B --> C[Не короть пины]'
WHERE lesson_id = 'b0000003-0000-0000-0000-000000000001'
  AND url_or_path = '/schematics/safety.svg';

UPDATE lesson_materials
SET kind = 'mermaid',
    url_or_path = E'flowchart TB\n  T[Thonny — среда для MicroPython]\n  T --> W[Скачать с thonny.org]\n  T --> F[Залить прошивку .uf2 на Pico]'
WHERE lesson_id = 'b0000004-0000-0000-0000-000000000001'
  AND url_or_path = '/schematics/thonny-ide.svg';

UPDATE lesson_materials
SET kind = 'mermaid',
    url_or_path = E'flowchart LR\n  GP[GP25] --> R[220 Ом]\n  R --> LED[Светодиод +]\n  LED --> GND[GND]'
WHERE lesson_id = 'b0000005-0000-0000-0000-000000000001'
  AND url_or_path = '/schematics/led-circuit.svg';

UPDATE lesson_materials
SET kind = 'mermaid',
    url_or_path = E'flowchart TB\n  P[Raspberry Pi Pico]\n  P --> GP[Пин GP25]\n  GP --> L[Встроенный светодиод]'
WHERE lesson_id = 'b0000006-0000-0000-0000-000000000001'
  AND url_or_path = '/schematics/pico-board.svg';
