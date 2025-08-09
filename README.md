# Argo parser

Argo is another command line arguments parser for Go.

## Why not standard `flags`?

- **Argo** supports parsing arguments from any strings and string slices, while flags can parse only program arguments. This is very useful for bots or interactive console apps;
- **Argo** provides user-friendly syntax for describing typed argument objects. This is easy to write and understand at first glance.

## Command line syntax

First, let's define what we're going to parse.

**Argo**'s input is **single line** of text (or string slice that represents it after joining items by spaces).

Input represents a set of:

- positional arguments,
- key-value pairs,
- binary flags (`true` if presented, `false` otherwise).

Example:

```
subscriptions get --user-id 123 --only-active -vt
```

In this example positional arguments "subscriptions" and "get", user-id parameter with the value 123 and flags "v" and "t" (whatever it means) are passed.

### Arg name format

Flags and parameters may be specified with short or long name. Short format is the name with single hyphen prefix like `-u`. Long format is the name with double-hyphen prefix like `--user-id`.

All the text after double hyphen up to the space character is considered as argument name, so you can use `--123`, `--@rgu:ment` and even `--"` as valid argument names (but please don't do it, I'm serious).

In case there are more than one character after single hyphen, every character up to the space character is considered as flag, so you can list flags with only one hyphen. So, `-abc` is three flags `a`, `b` and `c`. Obviously, you can't use this for key-value pairs since they must have some value for every key.

Every argument may appear more than one time. In this case, value will be an array. For example, for input `-k v1 -k v2 -k -v3` the value of `k` will be `[v1, v2, v3]`.

### A bit about string values

In case you need string value with spaces use quotation marks. You can use double (`"`), single (`'`) and backtick (`` ` ``) quotation marks.

Of course, you can escape it with backslash (`\`) as well as backslash itself.

Example:

```
search user --name "Aleksandr Markov" --extra '{"this is backslash": "\\"}' --extra2 "{\"escaping\": \"example\"}"
```

## Usage

Arguments are described as a tagged struct.

### Supported types

By the way, it seems like a nice opportunity for contribution.

- `string`
- `int`
- slice of one of the types above
- `bool` (for flags and only for them)

### Tags

Fields may be configured with `arg` tag with comma-separated options.

#### Specifying argument keys

Use single- or double-hyphen options for specifying argument key.

Some valid examples:

- ``UserID string `arg:"--user-id"` `` - only long name is specified;
- ``UserID string `arg:"-u"` `` - short name is specified explicitly, long one will be calculated implicitly;
- ``UserID string `arg:"--user-id,-u"` `` - both options are specified explicitly.

In the case long name is not explicitly specified with `arg` tag it exists anyway and calculated according to the following rules:

- Field name is separated by capital letters.
- Consequent items conaining only one letter are joined; 
- The first letter of every result's item are replaced to lowercase version of it;
- Result is joined by hyphen.

For example for `FieldName` field long argument name is implicitly specified as `--field-name`. For UserID due to the second step the result is `--user-id`. Anyway, we recommend explicitly specified names.

#### Marking a field as required

There is `required` tag for marking the field as required. Parser validates the field is passed and returns an error otherwise.

Let's modify the example above and make UserID field requried:

```
UserID string `arg:"--user-id,required"`
```

#### Handling positional arguments

There's `positional` tag for handling positional arguments. Add it to `[]string` field of your structure to store all positional arguments there.

But there is also another use case. Let's take a look at the following struct:

```
type MyFancyParams struct {
    Module string `arg:"positional"`
    Command string `arg:"positional"`
    Arguments []string `arg:"positional"`
}
```

**The first** positional argument will be set to **the first** field marked with `positional` tag. Same for the second argument. Another positional arguments will be stored to `Arguments` field since its type is `[]string` and it has `positional` tag.

Technically:

- `positional`-tagged fields may be of any supported "scalar" type (string, int);
- Positional arguments are stored to the `positional`-tagged fields in the same order as they are listed in the struct;
- Positional arguments are always required;
- "Default" field for storing unspecified positional arguments is always `[]string` with `positional` tag;
- The field can't be positional and hypen-named at the same time.

### Parsing examples

TBD

## Contribution

Feel free to open issues and pull requests.

Short list of rules:

- for bugs add example for reproduction;
- for new features add description, containing this info:
    - what problem it solves,
    - why the problem is being solved this way,
    - why it may be useful for other users;
- both for bugs and new features add tests (or, please, explicitly describe why it's not required);
- in case of adding new features it's a good practice to open issue with RFC first, then implement it and open pull request (it may save your time);
- don't forget to keep documentation (this README) up to date.
