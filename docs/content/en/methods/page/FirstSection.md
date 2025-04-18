---
title: FirstSection
description: Returns the Page object of the top-level section of which the given page is a descendant.
categories: []
keywords: []
params:
  functions_and_methods:
    returnType: page.Page
    signatures: [PAGE.FirstSection]
---

{{% glossary-term section %}}

> [!note]
> When called on the home page, the `FirstSection` method returns the `Page` object of the home page itself.

Consider this content structure:

```text
content/
├── auctions/
│   ├── 2023-11/
│   │   ├── _index.md     <-- first section: auctions
│   │   ├── auction-1.md
│   │   └── auction-2.md  <-- first section: auctions
│   ├── 2023-12/
│   │   ├── _index.md     
│   │   ├── auction-3.md
│   │   └── auction-4.md
│   ├── _index.md         <-- first section: auctions
│   ├── bidding.md
│   └── payment.md        <-- first section: auctions
├── books/
│   ├── _index.md         <-- first section: books
│   ├── book-1.md
│   └── book-2.md         <-- first section: books
├── films/
│   ├── _index.md         <-- first section: films
│   ├── film-1.md
│   └── film-2.md         <-- first section: films
└── _index.md             <-- first section: home
```

To link to the top-level section of which the current page is a descendant:

```go-html-template
<a href="{{ .FirstSection.RelPermalink }}">{{ .FirstSection.LinkTitle }}</a>
```
