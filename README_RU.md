# hypr-dock
### Интерактивная док-панель для Hyprland

<img width="1360" height="768" alt="250725_16h02m52s_screenshot" src="https://github.com/user-attachments/assets/041d2cf6-13ba-4c89-a960-1903073ff2d4" />
<img width="1360" height="768" alt="250725_16h03m09s_screenshot" src="https://github.com/user-attachments/assets/0c1ad8ca-37c1-4fd6-a48d-46f74c2d2609" />

[![YouTube](https://img.shields.io/badge/YouTube-Видео-FF0000?logo=youtube)](https://youtu.be/HHUZWHfNAl0?si=ZrRv2ggnPBEBS5oY)
[![AUR](https://img.shields.io/badge/AUR-Package-1793D1?logo=arch-linux)](https://aur.archlinux.org/packages/hypr-dock)

## Установка

### Зависимости

- `go` (make)
- `gtk3`
- `gtk-layer-shell`

### Установка
! Первая сборка может занимать крайне много времени из за привязки gtk3 !
```bash
git clone https://github.com/lotos-linux/hypr-dock.git
cd hypr-dock
make get
make build
make install
```

### Удаление
```bash
make uninstall
```

### Локальный запуск (dev mode)
```bash
make exec
```

## Запуск

### Параметры запуска:

```text
  -config string
    	config file (default "~/.config/hypr-dock")
  -dev
    	enable developer mode
  -log-level string
    	log level (default "info")
  -theme string
    	theme dir
```
#### Все параметры являются необязательными.

Конфигурация и темы по умолчания ставяться в `/etc/hypr-dock`
При первом запуске копируются в `~/.config/hypr-dock`
### Добавьте запуск в `hyprland.conf`:

```text
exec-once = hypr-dock
bind = Super, D, exec, hypr-dock
```

### И настройте блюр если он вам нужен
```text
layerrule = blur true,match:namespace hypr-dock
layerrule = ignore_alpha 0,match:namespace hypr-dock
layerrule = blur true,match:namespace dock-popup
layerrule = ignore_alpha 0,match:namespace dock-popup
```

#### Док поддерживает только один запущенный экземпляр, так что повторный запуск закроет предыдующий.

## Настройка

#### !Важно! 
Если у вас уже стоял `hypr-dock` или вы обновляетесь то прошу обратить внимание что начиная с версии 1.2.0 теперь используется `ini` подобный фомат конфигурации. И если у вас в пользовательской папке остались старые конфиги обязательно удалите их или переместите. Затем просто адаптируйте ваши настройки под новый формат

### В `hypr-dock.conf` доступны такие параметры

```ini
[General]
CurrentTheme = lotos

# Icon size (px) (default 23)
IconSize = 23

# Window overlay layer height (background, bottom, top, overlay) (default top)
Layer = top

# Exclusive Zone (true, false) (default true)
Exclusive = true

# SmartView (true, false) (default false)
SmartView = false

# Window position on screen (top, bottom, left, right) (default bottom)
Position = bottom

# Delay before hiding the dock (ms) (default 400)
AutoHideDelay = 400   # Only for SmartView

# Use system gap (true, false) (default true)
SystemGapUsed = true

# Indent from the edge of the screen (px) (default 8)
Margin = 8

# Distance of the context menu from the window (px) (default 5)
ContextPos = 5

[General.preview]
# Window thumbnail mode selection (none, live, static) (default none)
Mode = none
# "none"   - disabled (text menus)
# "static" - last window frame
# "live"   - window streaming
      
# !WARNING! 
# BY SETTING "Mode" TO "live" OR "static", YOU AGREE TO THE CAPTURE 
# OF WINDOW CONTENTS.
# THE "HYPR-DOCK" PROGRAM DOES NOT COLLECT, STORE, OR TRANSMIT ANY DATA.
# WINDOW CAPTURE OCCURS ONLY FOR THE DURATION OF THE THUMBNAIL DISPLAY!
#   
# Source code: https://github.com/lotos-linux/hypr-dock

# Live preview fps (0 - ∞) (default 30)
FPS = 30

# Live preview bufferSize (1 - 20) (default 5)
BufferSize = 5

# Popup show/hide/move delays (ms)
ShowDelay = 500  # (default 500)
HideDelay = 350  # (default 350)
MoveDelay = 100  # (default 100)
```
#### Если параметр не указан значение будет выставлено по умолчанию

## Разберем неочевидные параметры

### SmartView
Что то на подобии автоскрытия, если `true` то док находиться под всеми окнами, но если увести курсор мыши к краю экрана - док поднимается над ними

### Exclusive
Активирует особое поведение слоя при котором тайлинговые окна не перекрывают док

### SystemGapUsed
- При `SystemGapUsed = true` док будет задавать для себя отступ от края экрана беря значение из конфигурации `hyprland`, а конкретно значения `general:gaps_out`, при этом док динамически будет подхватывать изменение конфигурации `hyprland`
- При `SystemGapUsed = false` отступ от края экрана будет задаваться параметром `Margin`

### General.preview
- `ShowDelay`, `HideDelay`, `MoveDelay` - задержки действий попапа превью в милисекундах
- `FPS`, `BufferSize` - используются только при `Mode = live`


#### Настройки внешнего вида превью происходит через файлы темы



### Заклепленные приложения храняться в файле `~/.local/share/hypr-dock/pinned`
Для закрепления откройте контестное меню приложения в доке и нажмине `pin`/`unpin`
#### Например
```text
firefox
code-oss
kitty
org.telegram.desktop
nemo
org.kde.ark
sublime_text
qt6ct
one.ablaze.floorp
```
Вы можете менять его в ручную. Но зачем? ¯\_(ツ)_/¯

## Темы

#### Темы находяться в папке `~/.config/hypr-dock/themes/`

### Тема состоит из
- `theme.conf`
- `style.css`
- Папка с `svg` файлами для индикации количества запущенных приложения (смотрите [themes_RU.md](https://github.com/lotos-linux/hypr-dock/blob/main/docs/customize/themes_RU.md))

### Конфиг темы
```ini
[Theme]
# Distance between elements (px) (default 9)
Spacing = 5


[Theme.preview]
# Size (px) (default 120)
Size = 120

# Image/Stream border-radius (px) (default 0)
BorderRadius = 0

# Popup padding (px) (default 10)
Padding = 10
```
#### Файл `style.css` крутите как хотите. Позже сделаю подробную документацию по стилизации