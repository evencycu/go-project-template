# 4. ADR support tool

Date: 2020-06-22

## Status

Accepted

Amends [3. ADR format with lightweight ADR](0003-adr-format-with-lightweight-adr.md)

## Context

We need to provide some tool further reduce the effort of users.

## Decision

Use [adr-tools](https://github.com/npryce/adr-tools) and [adr-viewer](https://github.com/mrwilson/adr-viewer) to help

## Consequences

We can create related folders by `adr init`

And We can generate LADR template by `adr new adr format`

And we can supersede (in typo way) the number 2 decisions by `adr new -s 2 ADR format with lightweight ADR`

And we can amend the number 3 decisions by `adr new -l "3:Amends:Amended by" ADR support tool`

And we can launch web UI by `adr-viewer --serve` and access it via `localhost::8000`

Life is more easier
