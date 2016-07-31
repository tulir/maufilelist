# mauFileList
[![License](http://img.shields.io/:license-gpl3-blue.svg?style=flat-square)](http://www.gnu.org/licenses/gpl-3.0.html)

A program that generates configurable fancy file lists. Developed for [dl.maunium.net](https://dl.maunium.net) (<-- now also used as a live demo).

## Configuration
I'm too lazy to write a proper explanation for the config, since the [example](https://github.com/tulir293/maufilelist/blob/master/example/config.json) is mostly self-explanatory

## Files
mauFileList mainly uses two files to generate a file list. The format file is called [`.mfl-format.gohtml`](https://github.com/tulir293/maufilelist/blob/master/example/format.gohtml) and directory config is [`.mfl.json`](https://github.com/tulir293/maufilelist/blob/master/example/mfl.json).

When an user requests something, mauFileList searches for both of these files starting from the directory requested and checking for the two files in every directory between the requested one and the root directory.
The two files don't have to be in the same directory. It's quite common to have just one format file in the root and a config file for every directory.

There's also (or rather, there will be) `.mfl-override.json` which can be used to override the names and change timestamps of files in a folder. This file is only searched for in the directory the user requests.

### Format file
The format file uses the Go Template language mixed with HTML. There's an example in [example/format.gohtml](https://github.com/tulir293/maufilelist/blob/master/example/format.gohtml).

The data given to the template system contains three variables:
* `Directory` - The name specified in the directory config (or the name of the directory, if nothing was specified).
* `FieldNames` - The list of field names in the directory config.
* `Files` - A two-dimensional list of field data.

### Directory config
* `directory-name` (optional) - The name to give to the format file.
* `field-names` - The list of field names (also given directly to the format file.
* `enable-back-button` (NYI) - Whether or not a button to go up a directory should appear.
* `directory-list` - The parsing & field data instructions for directories.
* `file-list` - Same as `directory-list`, but for files.

#### `directory-list` and `file-list`
* `enabled` - Whether or not the specific instruction set should be enabled.
* `parsing` - List of regexes to use to parse file names. Files that don't match any of the regexes are ignored.
* `field-data` - List of instructions on how to fill the fields.

#### Field data instructions
There are three instruction types. You can combine all instruction types by simply entering them.
* Arguments
  * Values taken from capture groups in file name parsing regexes.
  * Referenced using a dollar sign (`$`) and the number of the capture group.
* Variables
  * (NYI) Some variables can be formatted by adding a semicolon (`;`) after the variable name and at the end of the formatting instructions.
  * `file-name` - The full name of the file.
  * `last-change` - The last change timestamp.
    * Can be formatted using the [Go date format](http://fuckinggodateformat.com).
  * (NYI) `file-size` - The size of the file.
    * (NYI) Can be formatted to specific sizes (`b`, `k`, `m`, `g`, `t`).
* Literal string
  * Wrap strings in backticks (`) to make them literal.
