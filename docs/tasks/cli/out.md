# **Спецификация текстового формата ответов**

## **1. Форматы ответов**

### **1.1 Успешный ответ с данными**
```
<ключ>: <значение>
<вложенный_ключ>:
    <подключ>: <значение>
```

**Пример:**
```
user:
    id: 42
    name: Ivan Petrov
    permissions:
        - read
        - write
```

### **1.2 Успешный ответ без данных**
```
ok: <сообщение>
```

**Пример:**
```
ok: settings updated successfully
```

### **1.3 Ответ с ошибкой**
```
error: <текст ошибки> [<код>]
```

**Пример:**
```
error: file not found [1]
```

## **2. Таблица стандартных форматов**

| Тип ответа          | Формат вывода                     | Примеры                          |
|---------------------|-----------------------------------|----------------------------------|
| Успех с данными     | Ключ-значение с табами            | `id: 42`<br>`name: Ivan`         |
| Успех без данных    | `ok: <сообщение>`                 | `ok: operation completed`        |
| Ошибка              | `error: <текст> [<код>]`          | `error: invalid input [2]`       |
| Вложенные данные    | С отступом 4 пробела              | `settings:`<br>`    theme: dark` |
| Списки значений     | С дефисом и отступом              | `- read`<br>`- write`            |

## **3. Особенности форматирования**

1. **Отступы**:
   - Каждый уровень вложенности: 4 пробела
   - Выравнивание значений после двоеточия

2. **Списки**:
   - Каждый элемент с новой строки
   - Префикс в виде дефиса и пробела

3. **Ошибки**:
   - Всегда начинаются с `error: `
   - Код ошибки в квадратных скобках

## **4. Примеры вывода**

### **Комплексные данные**
```
server:
    id: srv-01
    status: active
    ips:
        - 192.168.1.1
        - 10.0.0.1
    config:
        max_connections: 100
        timeout: 30
```

### **Простое подтверждение**
```
ok: cache cleared
```

### **Ошибка с кодом**
```
error: connection timeout [5]
```