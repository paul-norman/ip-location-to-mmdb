# IP Location To MMDB

This is simple CLI tool to convert any of the CSV datasets from the excellent [ip-location-db](https://github.com/sapics/ip-location-db) project into an MMDB file for faster lookups.

It's written in [Go](https://go.dev/) to allow it to compile to many platforms and run from a single binary.

## Usage

The tool is designed to accept a correctly formatted CSV file input and convert it to an MMDB file output. It has several options:

| Option         | Short | Description                                                                                                                                                                                        | Compulsory? |
|----------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| `-input`       | `-i`  | The path to the input CSV file *(relative or absolute)* <br/><br/>This may be called twice if IPV4 and IPV6 are to be combined.                                                                    | Yes         |
| `-output`      | `-o`  | The path to the output MMDB file *(relative or absolute)* <br/><br/>*If omitted, the input name will be used with the extension swapped*                                                           | No          |
| `-type`        | `-t`  | The type of record being converted: `country`, `asn` or `city`<br/><br/>*If omitted, the input name will be checked to see if it contains those words (or their plurals)*                          | No          |
| -`ipv`         | N/A   | The IP version that is being converted: `4` or `6`<br/><br/> *If omitted, the file name will be checked*                                                                                           | No          |
| -`record_size` | -r    | The MMDB [Record Size](https://github.com/maxmind/MaxMind-DB/blob/main/MaxMind-DB-spec.md): `24`, `28` or `32`<br/><br/> *If omitted, the file name will be checked and a sensible default chosen* | No          |
| -`compress`    | -z    | Gzip the output MMDB file for distribution                                                                                                                                                         | No          |

```Shell
ip-location-to-mmdb -i /path/to/input.csv -o /path/to/output.mmdb -t country -ipv 4 -r 24
```

Or if the files are named well *(as named in the project)*:

```Shell
ip-location-to-mmdb -i /path/to/dbip-country-ipv4.csv
```

To combine the IPV4 and IPV6 files:

```Shell
ip-location-to-mmdb -i /path/to/dbip-country-ipv4.csv -i /path/to/dbip-country-ipv6.csv -o /path/to/dbip-country-ipall.mmdb -t country -r 24
```

## Building

The tool can be built for all architectures using the command:

```Shell
make build
```

Or specific platforms using more specific commands:

```Shell
make build_linux
make build_darwin
make build_windows
```

Or specific platforms and architectures using even more specific commands:

```Shell
make build_linux_x64
make build_linux_arm64
make build_darwin_x64
make build_darwin_arm64
make build_windows_x64
make build_windows_arm64
```