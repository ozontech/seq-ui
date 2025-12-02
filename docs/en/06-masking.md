# Masking

Masking provides the ability to hide some of the information in the logs without modifying the logs themselves in the `seq-db`.

Masking applies to search, export, and aggregation operations.

See `handlers.seq_api.masking` section in [config](./02-configuration.md).

## Examples

### Simple

All log fields will be masked. The masks will be applied in sequence.

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'mask'
    - re: '@[a-z]+'
      mode: 'mask'
```

Before:
```json
{
  "message": "request from @host123",
  "user": "@ivan",
  "phone": "123-456-7890"
}
```

After:
```json
{
  "message": "request from ********",
  "user": "*****",
  "phone": "************"
}
```

### Process/Ignore fields

You can specify a list of fields that will be processed/ignored during masking.
The list can be either global for all masks, or local for each mask (local has the higher priority).

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'mask'
  process_fields:
    - 'private_phone'
```

Before:
```json
{
  "public_phone": "098-765-4321",
  "fake_phone": "123-456-7890",
  "private_phone": "123-456-7890"
}
```

After:
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

Before:
```json
{
  "public_phone": "098-765-4321",
  "fake_phone": "123-456-7890",
  "private_phone": "123-456-7890"
}
```

After:
```json
{
  "public_phone": "************",
  "fake_phone": "123-456-7890",
  "private_phone": "************"
}
```

### Groups

For partial masking, you must use the `groups` field.

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      groups: [1, 3]
      mode: 'mask'
```

Before:
```json
{
  "phone": "123-456-7890"
}
```

After:
```json
{
  "phone": "***-456-****"
}
```

### Mask modes

There are 3 masking modes: `mask`, `replace` and `cut`. The `mask` mode was used in the examples above. 

```yaml
masking:
  masks:
    - re: '(\d{3})-(\d{3})-(\d{4})'
      mode: 'replace'
      replace_word: <phone>
```

Before:
```json
{
  "phone": "123-456-7890"
}
```

After:
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

Before:
```json
{
  "message": "phone: 123-456-7890;"
}
```

After:
```json
{
  "message": "phone: ;"
}
```

## Field filters

Field filters provide the ability to apply masks only for those events whose fields fall under the filtering conditions.

### Field filter set

Field filter set is a set of filters that are interconnected by a logical condition (`or`, `and`, `not`).
Even if you need to apply only one filter, you must specify the `condition` field, but in this case it is ignored (except `not`).

```yaml
masking:
  masks:
    - ...
      field_filters:
        - condition: 'or'
          filters: [<filter1>, <filter2>, ...]
```

### Examples

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

Masked:
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

Not masked:
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

Masked:
```json
{
  "version": "1.23.4"
}
```

Not masked:
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