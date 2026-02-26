-- Форматирование кода в уроках «Умный дом на Pico W»: переводим inline-код в markdown code blocks

-- === УРОК 1: Pico W — выходим в интернет ===

UPDATE lesson_steps SET content = $$Для подключения к WiFi нужна библиотека `network` — она уже встроена в MicroPython, устанавливать не надо. Замени название сети и пароль на свои и запусти в Thonny.

```python
import network
import time

SSID = "Название_сети"    # Заменить на имя своей WiFi
PASSWORD = "Пароль"       # Заменить на свой пароль

# Создаём объект WiFi и включаем его
wlan = network.WLAN(network.STA_IF)
wlan.active(True)
wlan.connect(SSID, PASSWORD)

# Ждём подключения — обычно 5–15 секунд
while not wlan.isconnected():
    print("Подключаемся...")
    time.sleep(1)

# IP-адрес понадобится в следующем уроке — запомни его!
print("Подключено! IP:", wlan.ifconfig()[0])
```$$
WHERE lesson_id = 'd0000001-0000-0000-0000-000000000001' AND title = 'Подключаемся к WiFi';

-- === УРОК 2: Управляем светодиодом с телефона ===

UPDATE lesson_steps SET content = $$Длинная нога LED (плюс) — через резистор 220 Ом на GP15, короткая нога (минус) — на GND. Если хочешь обойтись без лишних деталей, используй встроенный светодиод:

```python
# Внешний LED на GP15
led = Pin(15, Pin.OUT)

# Или встроенный LED на плате (кавычки обязательны)
led = Pin("LED", Pin.OUT)
```$$
WHERE lesson_id = 'd0000002-0000-0000-0000-000000000002' AND title = 'Схема: LED на GP15';

UPDATE lesson_steps SET content = $$Полный код веб-сервера. Сначала подключаемся к WiFi, потом запускаем сокет и ждём запросы от браузера.

```python
import network, socket, time
from machine import Pin

# --- Подключение к WiFi ---
wlan = network.WLAN(network.STA_IF)
wlan.active(True)
wlan.connect("Твоя_сеть", "Твой_пароль")
while not wlan.isconnected():
    time.sleep(1)
print("IP:", wlan.ifconfig()[0])

# --- Настраиваем светодиод ---
led = Pin(15, Pin.OUT)

# --- Запускаем сервер на стандартном веб-порту 80 ---
addr = socket.getaddrinfo("0.0.0.0", 80)[0][-1]
s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(addr)
s.listen(1)

while True:
    cl, _ = s.accept()               # Ждём нового подключения
    request = cl.recv(1024).decode() # Читаем запрос браузера

    if "GET /on" in request:
        led.value(1)                 # Включаем LED
    elif "GET /off" in request:
        led.value(0)                 # Выключаем LED

    status = "Включён" if led.value() else "Выключен"

    # Отвечаем HTML-страницей с двумя кнопками
    html = (
        "<!DOCTYPE html><html><head><meta charset='utf-8'>"
        "<style>body{font-family:sans-serif;text-align:center;padding:20px}"
        "button{font-size:22px;padding:12px 28px;margin:8px}</style></head>"
        "<body><h1>Умный дом</h1>"
        "<p>Светодиод: " + status + "</p>"
        "<a href='/on'><button>Включить</button></a> "
        "<a href='/off'><button>Выключить</button></a>"
        "</body></html>"
    )
    cl.send("HTTP/1.0 200 OK\r\nContent-Type: text/html\r\n\r\n")
    cl.send(html)
    cl.close()
```$$
WHERE lesson_id = 'd0000002-0000-0000-0000-000000000002' AND title = 'Как работает код сервера';

-- === УРОК 3: Несколько кнопок ===

UPDATE lesson_steps SET content = $$Подключи три светодиода через резисторы 220 Ом: красный на GP13, жёлтый на GP14, зелёный на GP15. Все короткие ноги — на GND.

```python
from machine import Pin

red    = Pin(13, Pin.OUT)
yellow = Pin(14, Pin.OUT)
green  = Pin(15, Pin.OUT)
```$$
WHERE lesson_id = 'd0000003-0000-0000-0000-000000000003' AND title = 'Три LED на разных пинах';

UPDATE lesson_steps SET content = $$Каждая кнопка на странице ведёт по своему адресу: `/red/on`, `/red/off`, `/yellow/on` и так далее. В коде проверяем запрос и управляем нужным светодиодом:

```python
if "GET /red/on" in request:
    red.value(1)
elif "GET /red/off" in request:
    red.value(0)
elif "GET /yellow/on" in request:
    yellow.value(1)
elif "GET /yellow/off" in request:
    yellow.value(0)
elif "GET /green/on" in request:
    green.value(1)
elif "GET /green/off" in request:
    green.value(0)
```

В HTML делаем шесть кнопок — по две на каждый цвет.$$
WHERE lesson_id = 'd0000003-0000-0000-0000-000000000003' AND title = 'Разные адреса — разные команды';

UPDATE lesson_steps SET content = $$Перед формированием HTML читаем состояние каждого LED через `.value()` и подставляем в текст страницы:

```python
# Читаем текущий статус каждого LED
r_state = "Включён" if red.value()    else "Выключен"
y_state = "Включён" if yellow.value() else "Выключен"
g_state = "Включён" if green.value()  else "Выключен"

# Вставляем в HTML рядом с кнопками:
# "Красный: " + r_state + " ..."
```

Каждый раз, когда открываешь страницу или нажимаешь кнопку, видишь актуальный статус всех трёх LED.$$
WHERE lesson_id = 'd0000003-0000-0000-0000-000000000003' AND title = 'Показываем статус на странице';

UPDATE lesson_steps SET content = $$Добавь общую кнопку «Выключить всё» — она ведёт по адресу `/all/off`. По аналогии сделай `/all/on`.

```python
elif "GET /all/off" in request:
    red.value(0)
    yellow.value(0)
    green.value(0)

elif "GET /all/on" in request:
    red.value(1)
    yellow.value(1)
    green.value(1)
```$$
WHERE lesson_id = 'd0000003-0000-0000-0000-000000000003' AND title = 'Задание: кнопка «Выключить всё»';

-- === УРОК 4: Температура в браузере ===

UPDATE lesson_steps SET content = $$Библиотека `dht` уже встроена в MicroPython. Создаём объект датчика и читаем данные:

```python
import dht
from machine import Pin

sensor = dht.DHT22(Pin(15))

# Перед каждым чтением нужно вызвать measure() — это команда "измерь"
sensor.measure()

# Важно: между двумя вызовами measure() должно пройти не меньше 2 секунд
temp = sensor.temperature()  # Температура в °C
hum  = sensor.humidity()     # Влажность в %
print(temp, hum)
```$$
WHERE lesson_id = 'd0000004-0000-0000-0000-000000000004' AND title = 'Читаем данные с датчика';

UPDATE lesson_steps SET content = $$В цикле веб-сервера измеряем перед каждым ответом и подставляем данные в HTML:

```python
# Измеряем перед формированием страницы
sensor.measure()
temp = sensor.temperature()
hum  = sensor.humidity()

# Подставляем в HTML-строку
html = (
    "<!DOCTYPE html><html><head><meta charset='utf-8'></head><body>"
    "<h1>Температура в комнате</h1>"
    "<p>Температура: " + str(temp) + " C</p>"
    "<p>Влажность: "   + str(hum)  + " %</p>"
    "</body></html>"
)
```$$
WHERE lesson_id = 'd0000004-0000-0000-0000-000000000004' AND title = 'Данные в веб-браузере';

UPDATE lesson_steps SET content = $$Добавь тег `<meta http-equiv="refresh">` в раздел `<head>` — браузер будет сам перезагружать страницу каждые 5 секунд:

```html
<head>
  <meta charset="utf-8">
  <meta http-equiv="refresh" content="5">
  <!-- content="5" — обновлять каждые 5 секунд, можно изменить -->
</head>
```

Число можно уменьшить до 2 или увеличить до 10. Теперь не нужно нажимать «обновить» — данные появляются сами.$$
WHERE lesson_id = 'd0000004-0000-0000-0000-000000000004' AND title = 'Страница обновляется сама';

-- === УРОК 5: Умный ночник с WiFi ===

UPDATE lesson_steps SET content = $$Читаем освещённость через ADC и управляем LED по порогу:

```python
from machine import ADC, Pin

ldr = ADC(26)          # Фоторезистор на GP26
led = Pin(15, Pin.OUT)

# Значение от 0 до 65535: в темноте — больше, на свету — меньше
light = ldr.read_u16()
THRESHOLD = 40000      # Порог — подбери под свой датчик

if light > THRESHOLD:
    led.value(1)       # Темно — включаем ночник
else:
    led.value(0)       # Светло — выключаем
```

Этот код выполняется внутри цикла веб-сервера перед отправкой HTML.$$
WHERE lesson_id = 'd0000005-0000-0000-0000-000000000005' AND title = 'Код автоматического ночника';

UPDATE lesson_steps SET content = $$Добавляем в HTML текущее значение с фоторезистора и понятный статус:

```python
light  = ldr.read_u16()
status = "Темно — ночник включён" if light > THRESHOLD else "Светло — ночник выключен"

html = (
    "..."
    "<p>Освещённость: " + str(light) + " / 65535</p>"
    "<p>" + status + "</p>"
    "..."
)
```

Закрой датчик рукой или посвети фонариком — значение будет меняться в реальном времени. Так удобно подбирать правильный порог.$$
WHERE lesson_id = 'd0000005-0000-0000-0000-000000000005' AND title = 'Уровень освещённости в браузере';

UPDATE lesson_steps SET content = $$Добавляем переменную `override` и три кнопки: включить вручную, выключить вручную, вернуться в авторежим.

```python
override = None   # None = авторежим, 1 = принудительно вкл, 0 = принудительно выкл

# Обработка кнопок со страницы
if "GET /manual/on" in request:
    override = 1
elif "GET /manual/off" in request:
    override = 0
elif "GET /auto" in request:
    override = None          # Возвращаемся в авторежим

# Логика управления LED
if override is not None:
    led.value(override)      # Ручное управление
else:
    light = ldr.read_u16()
    led.value(1 if light > THRESHOLD else 0)  # Авто по датчику
```$$
WHERE lesson_id = 'd0000005-0000-0000-0000-000000000005' AND title = 'Ручное управление через страницу';

-- === УРОК 6: Охранная сигнализация ===

UPDATE lesson_steps SET content = $$Измеряем расстояние ультразвуком и сравниваем с порогом тревоги:

```python
from machine import Pin, time_pulse_us
import time

trig = Pin(17, Pin.OUT)
echo = Pin(18, Pin.IN)
buzz = Pin(16, Pin.OUT)

def measure_distance():
    # Посылаем короткий ультразвуковой импульс
    trig.value(0);  time.sleep_us(2)
    trig.value(1);  time.sleep_us(10)
    trig.value(0)
    # Измеряем время отражения и переводим в сантиметры
    duration = time_pulse_us(echo, 1, 30000)
    return duration / 2 / 29.1

ALARM_CM = 50   # Тревога, если объект ближе 50 см

dist = measure_distance()
if dist < ALARM_CM:
    buzz.value(1)   # Тревога!
else:
    buzz.value(0)   # Тихо
```$$
WHERE lesson_id = 'd0000006-0000-0000-0000-000000000006' AND title = 'Измеряем расстояние и задаём зону охраны';

UPDATE lesson_steps SET content = $$Добавляем флаг `armed` и управляем им через веб-страницу:

```python
armed = False   # False = охрана снята, True = на охране

# Обработка кнопок
if "GET /arm" in request:
    armed = True
elif "GET /disarm" in request:
    armed = False
    buzz.value(0)        # Снимаем тревогу при отключении охраны

# Логика сигнализации
dist  = measure_distance()
alarm = armed and (dist < ALARM_CM)

buzz.value(1 if alarm else 0)

# Статус для страницы
armed_str = "Активна" if armed else "Отключена"
alarm_str = "ТРЕВОГА!" if alarm else "Всё спокойно"
dist_str  = str(round(dist, 1)) + " см"
```$$
WHERE lesson_id = 'd0000006-0000-0000-0000-000000000006' AND title = 'Управление охраной через браузер';

-- === УРОК 7: Проектируем умную комнату ===

UPDATE lesson_steps SET content = $$Разбей программу на функции — так будет намного проще отлаживать и дополнять:

```python
import network, socket, time, dht
from machine import Pin, ADC

# --- Функция подключения к WiFi ---
def connect_wifi(ssid, password):
    wlan = network.WLAN(network.STA_IF)
    wlan.active(True)
    wlan.connect(ssid, password)
    while not wlan.isconnected():
        time.sleep(1)
    return wlan.ifconfig()[0]   # Возвращаем IP-адрес

# --- Функция чтения всех датчиков ---
def read_sensors():
    sensor.measure()
    return {
        "temp":  sensor.temperature(),
        "hum":   sensor.humidity(),
        "light": ldr.read_u16(),
        "dist":  measure_distance(),
    }

# --- Функция обработки запроса ---
def handle_request(request):
    # Здесь проверяем команды: if "GET /on" in request ...
    pass

# --- Функция формирования HTML ---
def make_html(data):
    # Здесь собираем HTML-строку из значений датчиков
    pass
```$$
WHERE lesson_id = 'd0000007-0000-0000-0000-000000000007' AND title = 'Структура программы';

-- === УРОК 8: Собираем умную комнату — схема ===

UPDATE lesson_steps SET content = $$Прежде чем запускать всё вместе, протестируй каждую часть отдельно короткой программой:

```python
# Тест 1: светодиод горит?
from machine import Pin
led = Pin(13, Pin.OUT)
led.value(1)

# Тест 2: DHT22 читает данные?
import dht
s = dht.DHT22(Pin(15))
s.measure()
print(s.temperature(), s.humidity())   # Должны появиться числа

# Тест 3: фоторезистор реагирует на свет?
from machine import ADC
print(ADC(26).read_u16())  # Закрой рукой — число должно вырасти

# Тест 4: датчик расстояния работает?
# Запусти measure_distance() из урока 6 и посмотри на числа в консоли
```

Так проще найти проблему — сначала проверяй каждый компонент, потом объединяй.$$
WHERE lesson_id = 'd0000008-0000-0000-0000-000000000008' AND title = 'Тестируем каждый компонент отдельно';

-- === УРОК 9: Собираем умную комнату — программа ===

UPDATE lesson_steps SET content = $$Структура полной программы — всё в одном файле. Открой новый файл в Thonny и собирай по частям:

```python
import network, socket, time, dht
from machine import Pin, ADC

# 1. Объявляем все устройства
led    = Pin(15, Pin.OUT)
buzz   = Pin(16, Pin.OUT)
trig   = Pin(17, Pin.OUT)
echo   = Pin(18, Pin.IN)
ldr    = ADC(26)
sensor = dht.DHT22(Pin(21))   # Выбери свободный пин

# 2. Глобальные переменные состояния
armed    = False
override = None
THRESHOLD = 40000
ALARM_CM  = 50

# 3. Подключаемся к WiFi
wlan = network.WLAN(network.STA_IF)
wlan.active(True)
wlan.connect("Твоя_сеть", "Твой_пароль")
while not wlan.isconnected():
    time.sleep(1)
print("IP:", wlan.ifconfig()[0])

# 4. Запускаем сервер
addr = socket.getaddrinfo("0.0.0.0", 80)[0][-1]
s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(addr)
s.listen(1)

# 5. Основной цикл
while True:
    cl, _ = s.accept()
    request = cl.recv(1024).decode()

    # Читаем датчики (handle_request и make_html — допиши сам!)
    sensor.measure()
    data = {
        "temp":  sensor.temperature(),
        "hum":   sensor.humidity(),
        "light": ldr.read_u16(),
    }

    cl.send("HTTP/1.0 200 OK\r\nContent-Type: text/html\r\n\r\n")
    cl.send(make_html(data))
    cl.close()
```$$
WHERE lesson_id = 'd0000009-0000-0000-0000-000000000009' AND title = 'Объединяем код из всех уроков';
