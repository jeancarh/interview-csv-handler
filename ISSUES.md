# Issues

## Problems found in the sample CSV

Line numbers count the header as line 1. Importing the file into an empty catalog gives **87 created, 5 skipped, 5 rejected** out of 97 rows.

| Line | SKU | Problem | What the import does |
| --- | --- | --- | --- |
| all | | The file uses Windows line endings (CRLF) | Parsed fine. The file is kept byte for byte in the repo (`.gitattributes` excludes it from line ending conversion) |
| 2 | RS-001 | Same SKU as line 36, with a different description and price | Skipped, line 36 wins |
| 4 | WM-042 | Price written as `$29.99` | Imported as 29.99, report notes the symbol was removed |
| 7 | YM-015 | Price is `free` | Rejected. Not guessed as 0, someone should confirm the price |
| 11 | BS-021 | Exact copy of line 89 | Skipped as a duplicate |
| 16 | DL-007 | Stock is `-5` | Rejected |
| 20 | XS-001 | Name is `<script>alert('xss')</script>` | Imported as plain text and shown escaped (you see the literal text, nothing runs). The report flags it: "name contains HTML markup" |
| 25 | HD-099 | Name is empty | Rejected |
| 29 | SQL-001 | Name is `Robert'); DROP TABLE products;--` | Imported as plain text, queries are parameterized |
| 31 | WB-033 | Description has `—` and `™` | Imported, UTF-8 is kept as is |
| 36 | RS-001 | Updated version of line 2 | Imported |
| 41 | WS-001 | Name is only spaces | Rejected |
| 47 | MB-001 | Price is `0.00` | Imported, a zero price is allowed |
| 50 | GK-088 | `weight_kg` is empty | Imported without weight, report says so |
| 51 | VC-001 | Stock is `0` | Imported, shows as sold out |
| 52 | GC-025 | Category is empty (stock `99999` and weight `0` are fine on their own) | Rejected |
| 53 | CI-001 | Comma inside the quoted name | Parsed correctly |
| 55 | RS-050 | Same name as RS-001, different SKU | Imported, names don't have to be unique |
| 56 | BS-021 | Same SKU as line 11, different description and price | Skipped, line 89 wins |
| 59 | QI-001 | Escaped quotes in the name | Parsed correctly |
| 62, 63 | | Completely empty rows | Skipped |
| 89 | BS-021 | Third row for this SKU, same as line 11 | Imported |

### A note on BS-021

The file has three rows for BS-021. Line 56 has different data, but line 89 repeats the original row from line 11. With "last row wins", the product ends up with the original data and line 56 is effectively discarded. That's consistent with the rule, but it's exactly the kind of case I'd raise with whoever produces the file. A conflicting row in the middle of the file usually means two sources wrote to the same export.
