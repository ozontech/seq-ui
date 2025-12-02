# Маскирование

Маскирование предоставляет возможность скрывать часть информации в логах без изменения самих логов в `seq-db`.

Маскирование применяется к операциям поиска, экспорта и агрегирования.

Смотрите раздел `handlers.seq_api.masking` в [конфигурации](./02-configuration.md).

## Примеры

### Простой

Все поля лога будут замаскированы. Маски будут применяться последовательно.

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'mask'
    - re: '@[a-z]+'
      mode: 'mask'
```

До:
```json
{
  "message": "request from @host123",
  "user": "@ivan",
  "phone": "123-456-7890"
}
```

После:
```json
{
  "message": "request from ********",
  "user": "*****",
  "phone": "************"
}
```

### Обрабатываемые/Игнорируемые поля

Вы можете указать список полей, которые будут обрабатываться/игнорироваться во время маскирования.
Список может быть как глобальным для всех масок, так и локальным для каждой отдельной маски (локальный имеет более высокий приоритет).

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'mask'
  process_fields:
    - 'private_phone'
```

До:
```json
{
  "public_phone": "098-765-4321",
  "fake_phone": "123-456-7890",
  "private_phone": "123-456-7890"
}
```

После:
```json
{
  "public_phone": "098-765-4321",
  "fake_phone": "123-456-7890",
  "private_phone": "************"
}
```

---

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'mask'
      ignore_fields:
        - 'fake_phone'
  process_fields:
    - 'fake_phone'
```

До:
```json
{
  "public_phone": "098-765-4321",
  "fake_phone": "123-456-7890",
  "private_phone": "123-456-7890"
}
```

После:
```json
{
  "public_phone": "************",
  "fake_phone": "123-456-7890",
  "private_phone": "************"
}
```

### Группы

Для частичного маскирования используется поле `groups`.

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      groups: [1, 3]
      mode: 'mask'
```

До:
```json
{
  "phone": "123-456-7890"
}
```

После:
```json
{
  "phone": "***-456-****"
}
```

### Режимы маскирования

Существует 3 режима маскирования: `mask`, `replace` и `cut`. В приведенных выше примерах использовался режим `mask`. 

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'replace'
      replace_word: <phone>
```

До:
```json
{
  "phone": "123-456-7890"
}
```

После:
```json
{
  "phone": "<phone>"
}
```

---

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'cut'
```

До:
```json
{
  "message": "phone: 123-456-7890;"
}
```

После:
```json
{
  "message": "phone: ;"
}
```

## Фильтрация по полям

Фильтрация по полям предоставляет возможность применять маски только для тех событий, поля которых попадают под условия фильтрации.

### Набор фильтров по полям

Набор фильтров по полям - это набор фильтров, которые связаны между собой логическим условием (`or`, `and`, `not`).
Даже если вам нужно применить только один фильтр, вы должны задать `condition`, но в этом случае оно игнорируется (за исключением `not`).

```yaml
masking:
  masks:
    - ...
      field_filters:
        - condition: 'or'
          filters: [<filter1>, <filter2>, ...]
```

### Примеры

```yaml
masking:
  masks:
    - ...
      field_filters:
        condition: 'or'
        filters:
          - filed: 'level'
            mode: 'equal'
            values: ['0', '1', '2', '3']
          - field: 'message'
            mode: 'contains'
            vaules: ['error', 'panic']
```

Будет замаскировано:
```json
{
  "level": "3",
  "message": "request failed"
}
```
```json
{
  "level": "6",
  "message": "parsing error occured"
}
```

Не будет замаскировано:
```json
{
  "level": "4",
  "message": "request failed"
}
```

---

```yaml
masking:
  masks:
    - ...
      field_filters:
        condition: 'not'
        filters:
          - filed: 'version'
            mode: 'suffix'
            values: ['test', 'rc']
```

Будет замаскировано:
```json
{
  "version": "1.23.4"
}
```

Не будет замаскировано:
```json
{
  "version": "1.23.4-test"
}
```
```json
{
  "version": "1.23.4-rc"
}
```