# `.mfl-override.json`
Ability to override file names and change dates while still having links to the original names.
Proposed format:
```json
{
  "format": "15:04:05 02.01.2006",
  "files": {
    "original-filename": {
      "name": "Overriden Filename",
      "timestamp": "14:54:58 20.07.2016"
    },
    "another-original-filename": {
      "name": "Foo",
      "timestamp": "20:12:00 20.12.2012"
    }
  }
}
```

# `enable-back-button` config option and formatting


# Variable formatting
Example:
`last-change;15:04:05 02.01.2006;`

# `file-size` variable
The length of the file. Formatted using a letter:
* `b` - bytes
* `k` - kilobytes
* `m` - megabytes
* `g` - gigabytes
* `t` - terabytes
