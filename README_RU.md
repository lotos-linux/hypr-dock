# hypr-dock
### Интерактовная док-панель для Hyprland

![screenshot1](https://github.com/user-attachments/assets/b98cdf7c-83b0-4c12-9da1-ada9e1543178)
![250318_10h19m33s_screenshot](https://github.com/user-attachments/assets/3ef014e4-4613-4e28-b186-71ce262db404)


https://github.com/user-attachments/assets/50d26918-ac74-4b3b-b384-9dd98c62a799



## Установка

### Зависимости

- `go` (make)
- `gtk3`
- `gtk-layer-shell`

### Установка
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

### Обновление
```bash
make update
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
-theme string
    theme dir (default "lotos")
```
#### Все параметры являются необязательными.

Конфигурация и темы по умолчания ставяться в `~/.config/hypr-dock`
### Добавьте запуск в `hyprland.conf`:

```text
exec-once = hypr-dock
bind = Super, D, exec, hypr-dock
```

#### Док поддерживает только один запущенный экземпляр, так что повторный запуск закроет предыдующий.

## Настройка

### В `config.jsonc` доступны такие параметры

```jsonc
{
    "CurrentTheme": "lotos",

    // Icon size (px) (default 23)
    "IconSize": 23,

    // Window overlay layer height (auto, exclusive-top, exclusive-bottom, background, bottom, top, overlay) (default "auto")
    "Layer": "auto",

    // Window position on screen (top, bottom, left, right) (default "bottom")
    "Position": "bottom",

    // Delay before hiding the dock (ms) (default 400)
    "AutoHideDeley": 400, // *Only for "Layer": "auto"*

    // Use system gap (true, false) (default "true")
    "SystemGapUsed": "true",

    // Indent from the edge of the screen (px) (default 8)
    "Margin": 8,

    // Distance of the context menu from the window (px) (default 0)
    "ContextPos": 5,

    
    // Window thumbnail mode selection (none, live, static) (default "none")
    "Preview": "none",
    /*
      "none"   - disabled (text menus)
      "static" - last window frame (stable)
      "live"   - window streaming (unstable) !EXPEREMENTAL!
      
      !WARNING! 
      BY SETTING "Preview" TO "live" OR "static", YOU AGREE TO THE CAPTURE 
      OF WINDOW CONTENTS.
      THE "HYPR-DOCK" PROGRAM DOES NOT COLLECT, STORE, OR TRANSMIT ANY DATA.
      WINDOW CAPTURE OCCURS ONLY FOR THE DURATION OF THE THUMBNAIL DISPLAY!
      
      Source code: https://github.com/lotos-linux/hypr-dock
    */

    "PreviewAdvanced": {
      // Live preview fps (0 - ∞) (default 30)
      "FPS": 30,

      // Live preview bufferSize (1 - 20) (default 5)
      "BufferSize": 5,

      // Popup show/hide/move delays (ms)
      "ShowDelay": 600, // (default 600)
      "HideDelay": 300, // (default 300)
      "MoveDelay": 200  // (default 200)
    }
}
```
#### Если параметр не указан значение будет выставлено по умолчанию

## Разберем неочевидные параметры
### Layer
- При `"Layer": "auto"` слой дока находиться под всеми окнами, но если увести курсор мыши к краю экрана - док поднимается над ними
- При `"Layer": "exclusive-top"` включается эксклюзивный режим на верхнем слое. Тайлинговые и плавающие окна не будут перекрывать док.
- При `"Layer": "exclusive-bottom"` включается эксклюзивный режим на нижнем слое. Тайлинговые окна не будут перекрывать док. Плавающие окна будут поверх дока.

### SystemGapUsed
- При `"SystemGapUsed": "true"` док будет задавать для себя отступ от края экрана беря значение из конфигурации `hyprland`, а конкретно значения `general:gaps_out`, при этом док динамически будет подхватывать изменение конфигурации `hyprland`
- При `"SystemGapUsed": "false"` отступ от края экрана будет задаваться параметром `Margin`

### PreviewAdvanced
- `ShowDelay`, `HideDelay`, `MoveDelay` - задержки действий попапа превью в милисекундах
- `FPS`, `BufferSize` - используются только при `"Preview":"live"`

> Внимание!
> Живое превью ведет себя не стабильно.
> Пока что не рекомендую ставить значение `"Preview": "live"`


#### Настройки внешнего вида превью происхрдит через файлы темы



### Также есть файл `pinned.json` с закрепленными приложениями
#### Например
```json
{
  "Pinned": [
    "firefox",
    "org.telegram.desktop",
    "code-oss",
    "kitty"
  ]
}
```
Вы можете менять его в ручную. Но зачем? ¯\_(ツ)_/¯

## Темы

#### Темы находяться в папке `~/.config/hypr-dock/themes/`

### Тема состоит из
- `[название_темы].jsonc` например `lotos.jsonc`
- `style.css`
- Папка с `svg` файлами для индикации количества запущенных приложения (смотрите [themes_RU.md](https://github.com/lotos-linux/hypr-dock/blob/main/docs/customize/themes_RU.md))

### Конфиг темы
```jsonc
{
    // Blur window ("true", "false") (default "on")
    "Blur": "true",

    // Distance between elements (px) (default 9)
    "Spacing": 5,

    // Preview settings
    "PreviewStyle": {
        // Size (px) (default 120)
        "Size": 120,

        // Image/Stream border-radius (px) (default 0)
        "BorderRadius": 0,

        // Popup padding (px) (default 10)
        "Padding": 10
    }
}
```
#### Файл `style.css` крутите как хотите 

## Использованные библиотки
- [github.com/akshaybharambe14/go-jsonc](https://github.com/akshaybharambe14/go-jsonc) v1.0.0
- [github.com/allan-simon/go-singleinstance](https://github.com/allan-simon/go-singleinstance) v0.0.0-20210120080615-d0997106ab37
- [github.com/dlasky/gotk3-layershell](https://github.com/dlasky/gotk3-layershell) v0.0.0-20240515133811-5c5115f0d774
- [github.com/goccy/go-json](https://github.com/goccy/go-json) v0.10.3
- [github.com/gotk3/gotk3](https://github.com/gotk3/gotk3) v0.6.3