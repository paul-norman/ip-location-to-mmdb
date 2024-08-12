# IP Location To MMDB

This is simple CLI tool to convert any of the CSV datasets from the excellent [ip-location-db](https://github.com/sapics/ip-location-db) project into an MMDB file for faster lookups.

It's written in [Go](https://go.dev/) to allow it to compile to many platforms and run from a single binary.

## Usage

The tool is designed to accept a correctly formatted CSV file input and convert it to an MMDB file output. It has several options:

| Option         | Short | Description                                                                                                                                                                                        | Compulsory? |
|----------------|-------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------|
| `-input`       | `-i`  | The path to the input CSV file *(relative or absolute)*                                                                                                                                            | Yes         |
| `-output`      | `-o`  | The path to the output MMDB file *(relative or absolute)* <br/><br/>*If omitted, the input name will be used with the extension swapped*                                                           | No          |
| `-type`        | `-t`  | The type of record being converted: `country`, `asn` or `city`<br/><br/>*If omitted, the input name will be checked to see if it contains those words (or their plurals)*                          | No          |
| -`ipv`         | N/A   | The IP version that is being converted: `4` or `6`<br/><br/> *If omitted, the file name will be checked*                                                                                           | No          |
| -`record_size` | -r    | The MMDB [Record Size](https://github.com/maxmind/MaxMind-DB/blob/main/MaxMind-DB-spec.md): `24`, `28` or `32`<br/><br/> *If omitted, the file name will be checked and a sensible default chosen* | No          |

```Shell
ip-location-to-mmdb -i /path/to/input.csv -o /path/to/output.csv -t country -ipv 4 -r 24
```

Or if the files are named well *(as named in the project)*:

```Shell
ip-location-to-mmdb -i /path/to/dbip-country-ipv4.csv
```