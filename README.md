# TL;DR

> This tool approaches analyzing Cairo smart contracts from two main ideas:
>
> 1. Uses tree-sitter to parse Cairo code and tree-sitter queries to find patterns.
> 2. Uses vector embeddings to adapt tree-sitter queries to the contract being analyzed
>    by replacing hardcoded identifiers with their semantic equivalents.

# 🧪 Solidair – Cairo Static Analysis for Protocol Vulnerabilities

A lightweight static analysis tool designed to detect patterns that led to the [zkLend hack](https://blocksec.com/blog/zklend-exploit-post-mortem) using Tree-sitter and Go.

---

## 🚨 Disclaimer

- I’m not trained as a security researcher, auditor or a Cairo expert.
- I'll try and approach things from first principles, document my reasoning and have fun.
- I'm not bouncing off any feedback, so a lot of ideas will likely be obvious, short-sighted, and may not survive much scrutiny. It is what it is :)
- Since this is a design excercise and not a programming exercise I leaned into AI generation to help me build.

---

## 🧠 Why zkLend?

I'm avoiding infrastructure hacks like stolen private keys. These often boil down to simply "failing at the basics". That just isn't fun. (EDIT: I read the article on private keys and how this takes an industry-wide
change. I agree.)

I focused on the zkLend hack (Feb 2025).

I've been briefly exposed to the aftermath, trying to help identify the attacker,
but I wasn't involved at all in any investigation. It's interesting, and fun, to revisit
this from a different angle.

---

## 🧵 What Happened in zkLend?

I see **three** main components that contributed to the hack:

1. **Integer floor division**

   - Mishandled even with a safe math library, likely due to misuse rather than a bug in the library.
   - Everyone’s already written about this, so I skipped it. I don't want to end up chasing
     an obvious idea and build a clone of existing work.

2. **Donation mechanics**

   - Overpaying flash loans boosts attacker balances.
   - Not too interesting for me to explore as part of this task.

3. **Uninitialized market conditions** ✅
   - The attack took advantage of a market that hadn’t been properly initialized.
   - This style of hack isn't super common, but when it happens, it tends to be rather costly.
     (e.g., Hundred Finance and Sonne Finance were multi-million-dollar hacks.)

---

## 🧰 Version 1

I don't want to build a full Cairo parser. I'll use tree-sitter instead.

Tree-sitter is a "parser generator tool and an incremental parsing library". It's typically
used in text editors (like Neovim) to provide syntax highlighting, code folding, etc.

We can leverage it for its query engine to find patterns in code. There is a grammar
definition for Cairo. Documentation is somewhat lacking, so it's not perfect but with
some effort, we can get the bindings to work in Golang.

This makes our design really, really simple. All our tool has to do is:

- take a Cairo file
- parse it with tree-sitter
- read query definitions from `queries/*.csm`
- parse the query metadata from the comments in the query. We can use these for reporting.
- run each query against the parsed tree and report matches

Example of a Tree-sitter query that checks for deposit-style functions that do not have
an indicator of active status:

```scm
(function_item
  (function
    name: (identifier) @func_name
    (#match? @func_name "^(deposit|provide_liquidity|add_liquidity|mint)$"))
  body: (block) @func_body
  (#not-match? @func_body "is_active|active|isActive|marketActive"))
```

```shell
$ solidair analyze examples/good.cairo
Loaded 8 queries
No vulnerabilities found.

$ solidair analyze examples/bad.cairo
Loaded 8 queries
Found 8 potential vulnerabilities:

Vulnerability #1: Missing Deposit Grace Period
Source: race_conditions/missing_deposit_grace_period.scm
Description: Deposit function lacks protection against front-running attacks
Line: 24
Code: deposit

Vulnerability #2: Missing Grace Period Variables
Source: race_conditions/missing_grace_period_variables.scm
Description: Contract doesn't have storage for grace period enforcement
Line: 8
Code: Storage
```

✅ Pros

- Fast (Tree-sitter is built for real-time parsing).
- Single binary, no deps (thanks to Go's embed).
- Easy to extend – just add .scm queries.
- We can add parsers for other languages (e.g. Solidity). This might need its own
  set of queries but otherwise wouldn't change the architecture of the tool.
- Better than regex, lighter than full semantic analysis.

❌ Cons

- Not semantic – it’s purely syntax-based. E.g. no types or scopes.
- Hardcoded identifiers – is_active, marketActive, etc.
- No rule composition – can’t AND/OR queries programmatically.

# 🧪 Version 2

In this version I want to tackle the problem with identifiers being hardcode.

The problem here is that we can't know up front how a smart contract writer will
name their variables.

We can do this by:

1. Extracting the variables used in the smart contract with a tree-sitter query.
2. We're going to calculate vector embeddings for those variable names.
3. We'll match those embeddings against our own embeddings.
4. We'll pass those matches as parameters to our queries.

So we have to do a few things:

- We need to build the functionality to extract variable names.
- We need functionality to calculate embeddings.
- We need to add support for treating our queries as templates so that we can
  pass values to them.
