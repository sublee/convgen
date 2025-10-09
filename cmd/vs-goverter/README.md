# Convgen vs. goverter

This directory contains code to compare
[Convgen](https://github.com/sublee/convgen) and
[goverter](https://github.com/jmattheis/goverter).

Generated code is already checked in. But you can regenerate it by running:

```bash
go run ./cmd/convgen ./cmd/vs-goverter
go run github.com/jmattheis/goverter/cmd/goverter@v1.9.2 gen ./cmd/vs-goverter
```

The program output is:

```
# Case 1: Successful conversion

Input:
  {ID:499602d2 Name:John Doe URLs:[https://example.com] Role:2 CreateTime:2025-01-02 03:04:05 +0000 UTC}
Convgen:
  {"id":"499602d2","firstname":"John","lastname":"Doe","urls":["https://example.com"],"role":"member","createdAt":1735787045}
Goverter:
  {"id":"499602d2","firstname":"John","lastname":"Doe","urls":["https://example.com"],"role":"member","createdAt":1735787045}

# Case 2: Comprehensive error message

Input:
  {ID:0 Name:Alice URLs:[] Role:0 CreateTime:0001-01-01 00:00:00 +0000 UTC}
Convgen:
  converting User.Name: need two parts to parse firstname: "Alice"
Goverter:
  need two parts to parse firstname: "Alice"
```